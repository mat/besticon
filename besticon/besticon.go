// Package besticon includes functions
// finding icons for a given web site.
package besticon

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/golang/groupcache"

	// Load supported image formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "github.com/mat/besticon/v3/ico"

	"github.com/mat/besticon/v3/colorfinder"

	"golang.org/x/net/html/charset"
)

// Besticon is the main interface to the besticon package.
type Besticon struct {
	httpClient *http.Client
	iconCache  *groupcache.Group
	logger     Logger

	defaultFormats      []string
	discardImageBytes   bool
	maxResponseBodySize int64
}

// New returns a new Besticon instance.
func New(opts ...Option) *Besticon {
	b := &Besticon{}

	for _, opt := range opts {
		opt.applyOption(b)
	}

	if len(b.defaultFormats) == 0 {
		b.defaultFormats = []string{"gif", "ico", "jpg", "png"}
	}

	if b.maxResponseBodySize == 0 {
		b.maxResponseBodySize = 10485760 // 10MB
	}

	if b.httpClient == nil {
		b.httpClient = NewDefaultHTTPClient()
	}

	if b.logger == nil {
		b.logger = NewDefaultLogger(os.Stdout)
	}

	return b
}

// Icon holds icon information.
type Icon struct {
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Format    string `json:"format"`
	Bytes     int    `json:"bytes"`
	Error     error  `json:"error"`
	Sha1sum   string `json:"sha1sum"`
	ImageData []byte `json:",omitempty"`
}

type IconFinder struct {
	b *Besticon

	FormatsAllowed  []string
	HostOnlyDomains []string
	icons           []Icon
}

func (b *Besticon) NewIconFinder() *IconFinder {
	return &IconFinder{
		b: b,
	}
}

func (f *IconFinder) FetchIcons(url string) ([]Icon, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
		url = "http://" + url
	}

	url = f.stripIfNecessary(url)

	var err error

	if f.b.CacheEnabled() {
		f.icons, err = f.b.resultFromCache(url)
	} else {
		f.icons, err = f.b.fetchIcons(url)
	}

	return f.Icons(), err
}

// stripIfNecessary removes everything from URL but the Scheme and Host
// part if URL.Host is found in HostOnlyDomains.
// This can be used for very popular domains like youtube.com where throttling is
// an issue.
func (f *IconFinder) stripIfNecessary(URL string) string {
	u, e := url.Parse(URL)
	if e != nil {
		return URL
	}

	for _, h := range f.HostOnlyDomains {
		if h == u.Host || h == "*" {
			domainOnlyURL := url.URL{Scheme: u.Scheme, Host: u.Host}
			return domainOnlyURL.String()
		}
	}

	return URL
}

func (f *IconFinder) IconInSizeRange(r SizeRange) *Icon {
	icons := f.Icons()

	// 1. SVG always wins
	for _, ico := range icons {
		if ico.Format == "svg" {
			return &ico
		}
	}

	// 2. Try to return smallest in range perfect..max
	sortIcons(icons, false)
	for _, ico := range icons {
		if (ico.Width >= r.Perfect && ico.Height >= r.Perfect) && (ico.Width <= r.Max && ico.Height <= r.Max) {
			return &ico
		}
	}

	// 3. Try to return biggest in range perfect..min
	sortIcons(icons, true)
	for _, ico := range icons {
		if (ico.Width >= r.Min && ico.Height >= r.Min) && (ico.Width <= r.Perfect && ico.Height <= r.Perfect) {
			return &ico
		}
	}

	return nil
}

func (f *IconFinder) MainColorForIcons() *color.RGBA {
	return MainColorForIcons(f.icons)
}

func (f *IconFinder) Icons() []Icon {
	return f.b.discardUnwantedFormats(f.icons, f.FormatsAllowed)
}

func (ico *Icon) Image() (*image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(ico.ImageData))
	return &img, err
}

func (b *Besticon) discardUnwantedFormats(icons []Icon, wantedFormats []string) []Icon {
	formats := b.defaultFormats
	if len(wantedFormats) > 0 {
		formats = wantedFormats
	}

	return filterIcons(icons, func(ico Icon) bool {
		return includesString(formats, ico.Format)
	})
}

type iconPredicate func(Icon) bool

func filterIcons(icons []Icon, pred iconPredicate) []Icon {
	var result []Icon
	for _, ico := range icons {
		if pred(ico) {
			result = append(result, ico)
		}
	}
	return result
}

func includesString(arr []string, str string) bool {
	for _, e := range arr {
		if e == str {
			return true
		}
	}
	return false
}

func (b *Besticon) fetchIcons(siteURL string) ([]Icon, error) {
	var links []string

	html, urlAfterRedirect, e := b.fetchHTML(siteURL)
	if e == nil {
		// Search HTML for icons
		links, e = findIconLinks(urlAfterRedirect, html)
		if e != nil {
			return nil, e
		}
	} else {
		// Unable to fetch the response or got a bad HTTP status code. Try default
		// icon paths. https://github.com/mat/besticon/discussions/47
		links, e = defaultIconURLs(siteURL)
		if e != nil {
			return nil, e
		}
	}

	icons := b.fetchAllIcons(links)
	icons = rejectBrokenIcons(icons)
	sortIcons(icons, true)

	return icons, nil
}

func (b *Besticon) fetchHTML(url string) ([]byte, *url.URL, error) {
	r, e := b.Get(url)
	if e != nil {
		return nil, nil, e
	}

	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		return nil, nil, errors.New("besticon: not found")
	}

	body, e := b.GetBodyBytes(r)
	if e != nil {
		return nil, nil, e
	}
	if len(body) == 0 {
		return nil, nil, errors.New("besticon: empty response")
	}

	reader := bytes.NewReader(body)
	contentType := r.Header.Get("Content-Type")
	utf8reader, e := charset.NewReader(reader, contentType)
	if e != nil {
		return nil, nil, e
	}
	utf8bytes, e := io.ReadAll(utf8reader)
	if e != nil {
		return nil, nil, e
	}

	return utf8bytes, r.Request.URL, nil
}

func MainColorForIcons(icons []Icon) *color.RGBA {
	if len(icons) == 0 {
		return nil
	}

	var icon *Icon
	// Prefer gif, jpg, png
	for _, ico := range icons {
		if ico.Format == "gif" || ico.Format == "jpg" || ico.Format == "png" {
			icon = &ico
			break
		}
	}
	// Try .ico else
	if icon == nil {
		for _, ico := range icons {
			if ico.Format == "ico" {
				icon = &ico
				break
			}
		}
	}

	if icon == nil {
		return nil
	}

	img, err := icon.Image()
	if err != nil {
		return nil
	}

	cf := colorfinder.ColorFinder{}
	mainColor, err := cf.FindMainColor(*img)
	if err != nil {
		return nil
	}

	return &mainColor
}

// Construct default icon URLs. A fallback if we can't fetch the HTML.
func defaultIconURLs(siteURL string) ([]string, error) {
	baseURL, e := url.Parse(siteURL)
	if e != nil {
		return nil, e
	}

	var links []string
	for _, path := range iconPaths {
		absoluteURL, e := absoluteURL(baseURL, path)
		if e != nil {
			return nil, e
		}
		links = append(links, absoluteURL)
	}

	return links, nil
}

func (b *Besticon) fetchAllIcons(urls []string) []Icon {
	ch := make(chan Icon)

	for _, u := range urls {
		go func(u string) { ch <- b.fetchIconDetails(u) }(u)
	}

	var icons []Icon
	for range urls {
		icon := <-ch
		icons = append(icons, icon)
	}
	return icons
}

func (b *Besticon) fetchIconDetails(url string) Icon {
	i := Icon{URL: url}

	response, e := b.Get(url)
	if e != nil {
		i.Error = e
		return i
	}

	body, e := b.GetBodyBytes(response)
	if e != nil {
		i.Error = e
		return i
	}

	if isSVG(body) {
		// Special handling for svg, which golang can't decode with
		// image.DecodeConfig. Fill in an absurdly large width/height so SVG always
		// wins size contests.
		i.Format = "svg"
		i.Width = 9999
		i.Height = 9999
	} else {
		cfg, format, e := image.DecodeConfig(bytes.NewReader(body))
		if e != nil {
			i.Error = fmt.Errorf("besticon: unknown image format: %s", e)
			return i
		}

		// jpeg => jpg
		if format == "jpeg" {
			format = "jpg"
		}

		i.Width = cfg.Width
		i.Height = cfg.Height
		i.Format = format
	}

	i.Bytes = len(body)
	i.Sha1sum = sha1Sum(body)
	if !b.discardImageBytes {
		i.ImageData = body
	}

	return i
}

// SVG detector. We can't use image.RegisterFormat, since RegisterFormat is
// limited to a simple magic number check. It's easy to confuse the first few
// bytes of HTML with SVG.
func isSVG(body []byte) bool {
	// is it long enough?
	if len(body) < 10 {
		return false
	}

	// does it start with something reasonable?
	switch {
	case bytes.Equal(body[0:2], []byte("<!")):
	case bytes.Equal(body[0:2], []byte("<?")):
	case bytes.Equal(body[0:4], []byte("<svg")):
	default:
		return false
	}

	// is there an <svg in the first 300 bytes?
	if off := bytes.Index(body, []byte("<svg")); off == -1 || off > 300 {
		return false
	}

	return true
}

func absoluteURL(baseURL *url.URL, path string) (string, error) {
	u, e := url.Parse(path)
	if e != nil {
		return "", e
	}

	u.Scheme = baseURL.Scheme
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	if u.Host == "" {
		u.Host = baseURL.Host
	}
	return baseURL.ResolveReference(u).String(), nil
}

func urlFromBase(baseURL *url.URL, path string) string {
	u := *baseURL
	u.Path = path
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	return u.String()
}

func rejectBrokenIcons(icons []Icon) []Icon {
	var result []Icon
	for _, img := range icons {
		if img.Error == nil && (img.Width > 1 && img.Height > 1) {
			result = append(result, img)
		}
	}
	return result
}

func sha1Sum(b []byte) string {
	hash := sha1.New()
	hash.Write(b)
	bs := hash.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

var BuildDate string // set via ldflags on Make
