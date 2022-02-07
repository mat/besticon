package besticon

import (
	"io"
	"log"
	"net/http"
	"time"
)

type Logger interface {
	LogError(err error)
	// LogResponse is called when an HTTP request has been executed. The duration is the time it took to execute the
	// request. When error is nil, the response is the response object. Otherwise, the response is nil.
	LogResponse(req *http.Request, resp *http.Response, duration time.Duration, err error)
}

func NewDefaultLogger(w io.Writer) Logger {
	return &defaultLogger{
		logger: log.New(w, "http:  ", log.LstdFlags|log.Lmicroseconds),
	}
}

var _ Logger = (*defaultLogger)(nil)

type defaultLogger struct {
	logger *log.Logger
}

func (d *defaultLogger) LogError(err error) {
	d.logger.Println("ERR:", err)
}

func (d *defaultLogger) LogResponse(req *http.Request, resp *http.Response, duration time.Duration, err error) {
	if err != nil {
		d.logger.Printf("Error: %s %s %s %.2fms",
			req.Method,
			req.URL,
			err,
			float64(duration)/float64(time.Millisecond),
		)
	} else {
		d.logger.Printf("%s %s %d %.2fms %d",
			req.Method,
			req.URL,
			resp.StatusCode,
			float64(duration)/float64(time.Millisecond),
			resp.ContentLength,
		)
	}
}
