package main

import (
	"expvar"
	"os"
	"runtime"
	"strconv"
	"time"
)

var (
	fetchCount  = expvar.NewInt("fetchCount")
	fetchErrors = expvar.NewInt("fetchErrors")
)

func init() {
	expvar.NewString("goVersion").Set(runtime.Version())
	expvar.NewString("iconVersion").Set(os.Getenv("GIT_REVISION"))

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
