package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	duration *prometheus.HistogramVec
)

func newPrometheusHandler(path string, f http.Handler) http.Handler {
	// return expvarHandler{counter: expvar.NewInt(path), handler: f}
	return promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"path": path}), f)
}

func init() {
	duration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"path", "method"},
	)
	prometheus.MustRegister(duration)
}
