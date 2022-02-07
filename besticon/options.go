package besticon

import (
	"net/http"
)

type Option interface {
	applyOption(b *Besticon)
}

type httpClientOption struct {
	client *http.Client
}

func (h *httpClientOption) applyOption(b *Besticon) {
	b.httpClient = h.client
}

// WithHTTPClient sets the http client to use for requests.
func WithHTTPClient(client *http.Client) Option {
	return &httpClientOption{
		client: client,
	}
}

type loggerOption struct {
	logger Logger
}

func (l *loggerOption) applyOption(b *Besticon) {
	b.logger = l.logger
}

// WithLogger sets the logger to use for logging.
func WithLogger(logger Logger) Option {
	return &loggerOption{
		logger: logger,
	}
}

type defaultFormatsOption struct {
	formats []string
}

func (d *defaultFormatsOption) applyOption(b *Besticon) {
	b.defaultFormats = d.formats
}

// WithDefaultFormats sets the default accepted formats.
func WithDefaultFormats(formats ...string) Option {
	return &defaultFormatsOption{
		formats: formats,
	}
}

type discardImageBytesOption struct {
	discardImageBytes bool
}

func (k *discardImageBytesOption) applyOption(b *Besticon) {
	b.discardImageBytes = k.discardImageBytes
}

// WithDiscardImageBytes sets whether to discard image bodies.
func WithDiscardImageBytes(discardImageBytes bool) Option {
	return &discardImageBytesOption{
		discardImageBytes: discardImageBytes,
	}
}
