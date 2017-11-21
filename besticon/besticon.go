// Package besticon includes functions
// finding icons for a given web site.
package besticon

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"image/color"

	// Load supported image formats.
	_ "image/gif"
	_ "image/png"

	_ "github.com/mat/besticon/ico"

	"github.com/mat/besticon/colorfinder"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html/charset"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

var defaultFormats []string

const MinIconSize = 0
const MaxIconSize = 500

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
	FormatsAllowed  []string
	HostOnlyDomains []string
	KeepImageBytes  bool
	icons           []Icon
}

func (f *IconFinder) FetchIcons(url string) ([]Icon, error) {
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "http:") && !strings.HasPrefix(url, "https:") {
		url = "http://" + url
	}

	url = f.stripIfNecessary(url)

	var err error

	if CacheEnabled() {
		f.icons, err = resultFromCache(url)
	} else {
		f.icons, err = fetchIcons(url)
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

	// Try to return smallest in range perfect..max
	sortIcons(icons, false)
	for _, ico := range icons {
		if (ico.Width >= r.Perfect && ico.Height >= r.Perfect) && (ico.Width <= r.Max && ico.Height <= r.Max) {
			return &ico
		}
	}

	// Try to return biggest in range perfect..min
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
	return discardUnwantedFormats(f.icons, f.FormatsAllowed)
}

func (ico *Icon) Image() (*image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(ico.ImageData))
	return &img, err
}

func discardUnwantedFormats(icons []Icon, wantedFormats []string) []Icon {
	formats := defaultFormats
	if len(wantedFormats) > 0 {
		formats = wantedFormats
	}

	return filterIcons(icons, func(ico Icon) bool {
		return includesString(formats, ico.Format)
	})
}

type iconPredicate func(Icon) bool

func filterIcons(icons []Icon, pred iconPredicate) []Icon {
	result := []Icon{}
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

func fetchIcons(siteURL string) ([]Icon, error) {
	html, url, e := fetchHTML(siteURL)
	if e != nil {
		return nil, e
	}

	links, e := findIconLinks(url, html)
	if e != nil {
		return nil, e
	}

	icons := fetchAllIcons(links)
	icons = rejectBrokenIcons(icons)
	sortIcons(icons, true)

	return icons, nil
}

const maxResponseBodySize = 10485760 // 10MB

func fetchHTML(url string) ([]byte, *url.URL, error) {
	r, e := get(url)
	if e != nil {
		return nil, nil, e
	}

	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		return nil, nil, errors.New("besticon: not found")
	}

	b, e := getBodyBytes(r)
	if e != nil {
		return nil, nil, e
	}
	if len(b) == 0 {
		return nil, nil, errors.New("besticon: empty response")
	}

	reader := bytes.NewReader(b)
	contentType := r.Header.Get("Content-Type")
	utf8reader, e := charset.NewReader(reader, contentType)
	if e != nil {
		return nil, nil, e
	}
	utf8bytes, e := ioutil.ReadAll(utf8reader)
	if e != nil {
		return nil, nil, e
	}

	return utf8bytes, r.Request.URL, nil
}

var iconPaths = []string{
	"/favicon.ico",
	"/apple-touch-icon.png",
	"/apple-touch-icon-precomposed.png",
}

type empty struct{}

func findIconLinks(siteURL *url.URL, html []byte) ([]string, error) {
	doc, e := docFromHTML(html)
	if e != nil {
		return nil, e
	}

	baseURL := determineBaseURL(siteURL, doc)
	links := make(map[string]empty)

	// Add common, hard coded icon paths
	for _, path := range iconPaths {
		links[urlFromBase(baseURL, path)] = empty{}
	}

	// Add icons found in page
	urls := extractIconTags(doc)
	for _, url := range urls {
		url, e := absoluteURL(baseURL, url)
		if e == nil {
			links[url] = empty{}
		}
	}

	// Turn unique keys into array
	result := []string{}
	for u := range links {
		result = append(result, u)
	}
	sort.Strings(result)

	return result, nil
}

func determineBaseURL(siteURL *url.URL, doc *goquery.Document) *url.URL {
	baseTagHref := extractBaseTag(doc)
	if baseTagHref != "" {
		baseTagURL, e := url.Parse(baseTagHref)
		if e != nil {
			return siteURL
		}
		return baseTagURL
	}

	return siteURL
}

func docFromHTML(html []byte) (*goquery.Document, error) {
	doc, e := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if e != nil || doc == nil {
		return nil, errParseHTML
	}
	return doc, nil
}

var csspaths = strings.Join([]string{
	"link[rel='icon']",
	"link[rel='shortcut icon']",
	"link[rel='apple-touch-icon']",
	"link[rel='apple-touch-icon-precomposed']",

	// Capitalized variants, TODO: refactor
	"link[rel='ICON']",
	"link[rel='SHORTCUT ICON']",
	"link[rel='APPLE-TOUCH-ICON']",
	"link[rel='APPLE-TOUCH-ICON-PRECOMPOSED']",
}, ", ")

var errParseHTML = errors.New("besticon: could not parse html")

func extractBaseTag(doc *goquery.Document) string {
	href := ""
	doc.Find("head base[href]").First().Each(func(i int, s *goquery.Selection) {
		href, _ = s.Attr("href")
	})
	return href
}

func extractIconTags(doc *goquery.Document) []string {
	hits := []string{}
	doc.Find(csspaths).Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if ok && href != "" {
			hits = append(hits, href)
		}
	})
	return hits
}

func MainColorForIcons(icons []Icon) *color.RGBA {
	if len(icons) == 0 {
		return nil
	}

	var icon *Icon
	// Prefer .png and .gif
	for _, ico := range icons {
		if ico.Format == "png" || ico.Format == "gif" {
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
	color, err := cf.FindMainColor(*img)
	if err != nil {
		return nil
	}

	return &color
}

func fetchAllIcons(urls []string) []Icon {
	ch := make(chan Icon)

	for _, u := range urls {
		go func(u string) { ch <- fetchIconDetails(u) }(u)
	}

	icons := []Icon{}
	for range urls {
		icon := <-ch
		icons = append(icons, icon)
	}
	return icons
}

func fetchIconDetails(url string) Icon {
	i := Icon{URL: url}

	response, e := get(url)
	if e != nil {
		i.Error = e
		return i
	}

	b, e := getBodyBytes(response)
	if e != nil {
		i.Error = e
		return i
	}

	cfg, format, e := image.DecodeConfig(bytes.NewReader(b))
	if e != nil {
		i.Error = fmt.Errorf("besticon: unknown image format: %s", e)
		return i
	}

	i.Width = cfg.Width
	i.Height = cfg.Height
	i.Format = format
	i.Bytes = len(b)
	i.Sha1sum = sha1Sum(b)
	if keepImageBytes {
		i.ImageData = b
	}

	return i
}

func get(urlstring string) (*http.Response, error) {
	u, e := url.Parse(urlstring)
	if e != nil {
		return nil, e
	}
	// Maybe we can get rid of this conversion someday
	// https://github.com/golang/go/issues/13835
	u.Host, e = idna.ToASCII(u.Host)
	if e != nil {
		return nil, e
	}

	req, e := http.NewRequest("GET", u.String(), nil)
	if e != nil {
		return nil, e
	}

	setDefaultHeaders(req)

	start := time.Now()
	resp, err := client.Do(req)
	end := time.Now()
	duration := end.Sub(start)

	if err != nil {
		logger.Printf("Error: %s %s %s %.2fms",
			req.Method,
			req.URL,
			err,
			float64(duration)/float64(time.Millisecond),
		)
	} else {
		logger.Printf("%s %s %d %.2fms %d",
			req.Method,
			req.URL,
			resp.StatusCode,
			float64(duration)/float64(time.Millisecond),
			resp.ContentLength,
		)
	}

	return resp, err
}

func getBodyBytes(r *http.Response) ([]byte, error) {
	limitReader := io.LimitReader(r.Body, maxResponseBodySize)
	b, e := ioutil.ReadAll(limitReader)
	r.Body.Close()

	if len(b) >= maxResponseBodySize {
		return nil, errors.New("body too large")
	}
	return b, e
}

func setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 10_0 like Mac OS X) AppleWebKit/602.1.38 (KHTML, like Gecko) Version/10.0 Mobile/14A5297c Safari/602.1")
}

func mustInitCookieJar() *cookiejar.Jar {
	options := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, e := cookiejar.New(&options)
	if e != nil {
		panic(e)
	}

	return jar
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	setDefaultHeaders(req)

	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	return nil
}

func absoluteURL(baseURL *url.URL, path string) (string, error) {
	url, e := url.Parse(path)
	if e != nil {
		return "", e
	}

	url.Scheme = baseURL.Scheme
	if url.Scheme == "" {
		url.Scheme = "http"
	}

	if url.Host == "" {
		url.Host = baseURL.Host
	}
	return baseURL.ResolveReference(url).String(), nil
}

func urlFromBase(baseURL *url.URL, path string) string {
	url := *baseURL
	url.Path = path
	if url.Scheme == "" {
		url.Scheme = "http"
	}

	return url.String()
}

func rejectBrokenIcons(icons []Icon) []Icon {
	result := []Icon{}
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

var client *http.Client
var keepImageBytes bool

func init() {
	setHTTPClient(&http.Client{Timeout: 20 * time.Second})

	// Needs to be kept in sync with those image/... imports
	defaultFormats = []string{"png", "gif", "ico"}
}

func setHTTPClient(c *http.Client) {
	c.Jar = mustInitCookieJar()
	c.CheckRedirect = checkRedirect
	client = c
}

var logger *log.Logger

// SetLogOutput sets the output for the package's logger.
func SetLogOutput(w io.Writer) {
	logger = log.New(w, "http:  ", log.LstdFlags|log.Lmicroseconds)
}

func init() {
	SetLogOutput(os.Stdout)
	keepImageBytes = true
}

var BuildDate string // set via ldflags on Make
