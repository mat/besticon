package main

import (
	"expvar"
	"os"
	"runtime"
	"time"
)

var (
	fetchCount  = expvar.NewInt("fetchCount")
	fetchErrors = expvar.NewInt("fetchErrors")
)

func init() {
	expvar.NewString("goVersion").Set(runtime.Version())
	expvar.NewString("iconVersion").Set(os.Getenv("GIT_REVISION"))

	expvar.NewString("timeStartup").Set(time.Now().String())
	expvar.Publish("timeCurrent", expvar.Func(func() interface{} { return time.Now() }))
}
