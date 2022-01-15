package besticon

import (
	"errors"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

var _ http.RoundTripper = (*httpTransport)(nil)

type httpTransport struct {
	transport http.RoundTripper

	userAgent string
}

func (h *httpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "*/*")
	req.Header.Set("User-Agent", h.userAgent)
	return h.transport.RoundTrip(req)
}

func NewDefaultHTTPTransport(userAgent string) http.RoundTripper {
	return &httpTransport{
		transport: http.DefaultTransport,
		userAgent: userAgent,
	}
}

func NewDefaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   5 * time.Second,
		Jar:       mustInitCookieJar(),
		Transport: NewDefaultHTTPTransport("Mozilla/5.0 (iPhone; CPU iPhone OS 10_0 like Mac OS X) AppleWebKit/602.1.38 (KHTML, like Gecko) Version/10.0 Mobile/14A5297c Safari/602.1"),
	}
}

func (b *Besticon) Get(urlstring string) (*http.Response, error) {
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

	start := time.Now()
	resp, err := b.httpClient.Do(req)
	end := time.Now()
	duration := end.Sub(start)

	b.logger.LogResponse(req, resp, duration, err)

	return resp, err
}

func (b *Besticon) GetBodyBytes(r *http.Response) ([]byte, error) {
	limitReader := io.LimitReader(r.Body, b.maxResponseBodySize)
	data, e := ioutil.ReadAll(limitReader)
	r.Body.Close()

	if int64(len(data)) >= b.maxResponseBodySize {
		return nil, errors.New("body too large")
	}
	return data, e
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
