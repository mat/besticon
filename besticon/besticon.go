// Package besticon includes functions
// finding icons for a given web site.
package besticon

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sort"
	"strings"
	"time"

	// Load support for common image formats.
	_ "code.google.com/p/go.image/bmp"
	_ "code.google.com/p/go.image/tiff"
	_ "code.google.com/p/go.image/webp"
	_ "github.com/mat/besticon/ico"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"code.google.com/p/go.net/html/charset"
	"code.google.com/p/go.net/publicsuffix"
	"github.com/PuerkitoBio/goquery"
)

// Icon holds icon information.
type Icon struct {
	URL     string `json:"url"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Format  string `json:"format"`
	Bytes   int    `json:"bytes"`
	Error   error  `json:"error"`
	Sha1sum string `json:"sha1sum"`
}

// FetchBestIcon takes a siteURL and returns the icon with
// the largest dimensions for this site or an error.
func FetchBestIcon(siteURL string) (*Icon, error) {
	return fetchBestIconWithClient(siteURL, &http.Client{})
}

func fetchBestIconWithClient(siteURL string, c *http.Client) (*Icon, error) {
	icons, e := fetchIconsWithClient(siteURL, c)
	if e != nil {
		return nil, e
	}

	if len(icons) < 1 {
		return nil, errors.New("besticon: no icons found for site")
	}

	best := icons[0]
	return &best, nil
}

// FetchIcons takes a siteURL and returns all icons for this site
// or an error.
func FetchIcons(siteURL string) ([]Icon, error) {
	c := &http.Client{Timeout: 60 * time.Second}
	return fetchIconsWithClient(siteURL, c)
}

// fetchIconsWithClient modifies c's checkRedirect and Jar!
func fetchIconsWithClient(siteURL string, c *http.Client) ([]Icon, error) {
	configureClient(c)

	siteURL = strings.TrimSpace(siteURL)
	if !strings.HasPrefix(siteURL, "http") {
		siteURL = "http://" + siteURL
	}

	html, url, e := fetchHTML(siteURL, c)
	if e != nil {
		return nil, e
	}

	links, e := assembleIconLinks(url, html)
	if e != nil {
		return nil, e
	}

	icons := fetchAllIcons(links, c)
	icons = rejectBrokenIcons(icons)

	// Order when finished: (width/height, bytes, url)
	sort.Stable(byURL(icons))
	sort.Stable(byBytes(icons))
	sort.Stable(sort.Reverse(byWidthHeight(icons)))

	return icons, nil
}

func fetchHTML(url string, c *http.Client) ([]byte, *url.URL, error) {
	r, e := get(c, url)
	if e != nil {
		return nil, nil, e
	}

	if !(r.StatusCode >= 200 && r.StatusCode < 300) {
		return nil, nil, errors.New("besticon: not found")
	}

	b, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return nil, nil, e
	}
	defer r.Body.Close()
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

func assembleIconLinks(siteURL *url.URL, html []byte) ([]string, error) {
	links := make(map[string]empty)

	// Add common, hard coded icon paths
	for _, path := range iconPaths {
		links[urlFromBase(siteURL, path)] = empty{}
	}

	// Add icons found in page
	urls, e := findIcons(html)
	if e != nil {
		return nil, e
	}
	for _, url := range urls {
		url, e := absoluteURL(siteURL, url)
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

var csspaths = strings.Join([]string{
	"link[rel='icon']",
	"link[rel='shortcut icon']",
	"link[rel='apple-touch-icon']",
	"link[rel='apple-touch-icon-precomposed']",
}, ", ")

var errParseHTML = errors.New("besticon: could not parse html")

func findIcons(html []byte) ([]string, error) {
	doc, e := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if e != nil || doc == nil {
		return nil, errParseHTML
	}

	hits := []string{}
	doc.Find(csspaths).Each(func(i int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if ok && href != "" {
			hits = append(hits, href)
		}
	})

	return hits, nil
}

func fetchAllIcons(urls []string, c *http.Client) []Icon {
	ch := make(chan Icon)

	for _, u := range urls {
		go func(u string) { ch <- fetchIconDetails(u, c) }(u)
	}

	icons := []Icon{}
	for range urls {
		icon := <-ch
		icons = append(icons, icon)
	}
	return icons
}

func fetchIconDetails(url string, c *http.Client) Icon {
	i := Icon{URL: url}

	response, e := get(c, url)
	if e != nil {
		i.Error = e
		return i
	}

	b, e := ioutil.ReadAll(response.Body)
	if e != nil {
		i.Error = e
		return i
	}
	defer response.Body.Close()

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

	return i
}

func get(client *http.Client, url string) (*http.Response, error) {
	req, e := http.NewRequest("GET", url, nil)
	if e != nil {
		return nil, e
	}

	setDefaultHeaders(req)
	return client.Do(req)
}

func setDefaultHeaders(req *http.Request) {
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_3) AppleWebKit/534.55.3 (KHTML, like Gecko) Version/5.1.3 Safari/534.53.10")
}

func configureClient(c *http.Client) {
	c.Jar = mustInitCookieJar()
	c.CheckRedirect = checkRedirect
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
	if url.Host == "" {
		url.Host = baseURL.Host
	}
	return url.String(), nil
}

func urlFromBase(baseURL *url.URL, path string) string {
	url := *baseURL
	url.Path = path
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

func (i *Icon) ImgWidth() int {
	return i.Width / 2.0
}

type byWidthHeight []Icon

func (a byWidthHeight) Len() int      { return len(a) }
func (a byWidthHeight) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byWidthHeight) Less(i, j int) bool {
	return (a[i].Width < a[j].Width) || (a[i].Height < a[j].Height)
}

type byBytes []Icon

func (a byBytes) Len() int           { return len(a) }
func (a byBytes) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byBytes) Less(i, j int) bool { return (a[i].Bytes < a[j].Bytes) }

type byURL []Icon

func (a byURL) Len() int           { return len(a) }
func (a byURL) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byURL) Less(i, j int) bool { return (a[i].URL < a[j].URL) }
