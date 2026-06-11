package besticon

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
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
		// Re-validate the target of every redirect hop. Without this, the
		// initial-host check in Get only covers the first request: a public
		// host could 302 to a private/loopback/link-local address and the
		// default redirect-following client would happily follow it.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return checkPublicHost(req.URL.Hostname())
		},
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

	if e := checkPublicHost(u.Hostname()); e != nil {
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

// checkPublicHost resolves host and rejects it if it maps to a
// loopback/private address. It is applied both to the initial URL and to the
// target of every redirect hop so that an allowed public host cannot be used
// to bounce a request onto an internal address.
func checkPublicHost(host string) error {
	ipAddr, e := net.ResolveIPAddr("ip", host)
	if e != nil {
		return e
	}
	if isPrivateIP(ipAddr) {
		return errors.New("private ip address disallowed")
	}
	return nil
}

func isPrivateIP(ipAddr *net.IPAddr) bool {
	if ipAddr == nil {
		return false
	}

	return ipAddr.IP.IsLoopback() || ipAddr.IP.IsPrivate()
}

func (b *Besticon) GetBodyBytes(r *http.Response) ([]byte, error) {
	limitReader := io.LimitReader(r.Body, b.maxResponseBodySize)
	data, e := io.ReadAll(limitReader)
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
