package main

import (
	"expvar"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/mat/besticon/besticon"
)

type expvarHandler struct {
	counter *expvar.Int
	handler http.Handler
}

func newExpvarHandler(path string, f http.Handler) expvarHandler {
	return expvarHandler{counter: expvar.NewInt(path), handler: f}
}

func (h expvarHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.counter.Add(1)
	h.handler.ServeHTTP(w, r)
}

func init() {
	expvar.NewString("goVersion").Set(runtime.Version())
	expvar.NewString("iconVersion").Set(besticon.VersionString)

	expvar.NewString("timeLastDeploy").Set(parseUnixTimeStamp(os.Getenv("DEPLOYED_AT")).String())
	expvar.NewString("timeStartup").Set(time.Now().String())
	expvar.Publish("timeCurrent", expvar.Func(func() interface{} { return time.Now() }))
}

func parseUnixTimeStamp(s string) time.Time {
	ts, err := strconv.Atoi(s)
	if err != nil {
		return time.Unix(0, 0)
	}

	return time.Unix(int64(ts), 0)
}
