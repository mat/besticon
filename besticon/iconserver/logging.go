package main

import (
	"log"
	"net/http"
	"os"
	"time"
)

var logger = log.New(os.Stdout, "server: ", log.LstdFlags|log.Lmicroseconds)

type loggingWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *loggingWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *loggingWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}

	bytesWritten, err := w.ResponseWriter.Write(b)
	if err == nil {
		w.length += bytesWritten
	}
	return bytesWritten, err
}

func newLoggingMux() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		writer := loggingWriter{w, 0, 0}
		http.DefaultServeMux.ServeHTTP(&writer, req)
		end := time.Now()
		duration := end.Sub(start)

		logger.Printf("%s %s %d \"%s\" %s %.2fms %d",
			req.Method,
			req.URL,
			writer.status,
			req.UserAgent(),
			req.Referer(),
			float64(duration)/float64(time.Millisecond),
			writer.length,
		)
	}
}
