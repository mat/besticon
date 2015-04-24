package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/mat/besticon/besticon"
	"github.com/mat/besticon/besticon/iconserver/assets"
)

func iconsHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue(urlParam)
	if len(url) == 0 {
		http.Redirect(w, r, "/", 302)
		return
	}

	icons, e := fetchIcons(url)
	switch {
	case e != nil:
		renderHTMLTemplate(w, 404, iconsHTML, pageInfo{URL: url, Error: e})
	case len(icons) <= 0:
		errNoIcons := errors.New("this poor site has no icons at all :-(")
		renderHTMLTemplate(w, 404, iconsHTML, pageInfo{URL: url, Error: errNoIcons})
	default:
		renderHTMLTemplate(w, 200, iconsHTML, pageInfo{Icons: icons, URL: url})
	}
}

const urlParam = "url"
const bestParam = "i_am_feeling_lucky"

func apiHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue(urlParam)
	if len(url) == 0 {
		errMissingURL := errors.New("need url query parameter")
		writeAPIError(w, 400, errMissingURL)
		return
	}

	bestMode := r.FormValue(bestParam) == "yes"
	if bestMode {
		icon, e := fetchBestIcon(url)
		if e != nil {
			writeAPIError(w, 404, e)
			return
		}

		http.Redirect(w, r, icon.URL, 302)
	} else {
		icons, e := fetchIcons(url)
		if e != nil {
			writeAPIError(w, 404, e)
			return
		}

		writeAPIIcons(w, url, icons)
	}
}

func fetchIcons(url string) ([]besticon.Icon, error) {
	fetchCount.Add(1)
	icons, err := besticon.FetchIcons(url)
	if err != nil {
		fetchErrors.Add(1)
	}
	return icons, err
}

func fetchBestIcon(url string) (*besticon.Icon, error) {
	fetchCount.Add(1)
	icon, err := besticon.FetchBestIcon(url)
	if err != nil {
		fetchErrors.Add(1)
	}
	return icon, err
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	stats := stats{}
	stats.Stats = append(stats.Stats, pair{"Current Time", time.Now().String()})
	stats.Stats = append(stats.Stats, pair{"Last Deploy", parseUnixTimeStamp(os.Getenv("DEPLOYED_AT")).String()})
	stats.Stats = append(stats.Stats, pair{"Deployed Git Revision", os.Getenv("GIT_REVISION")})
	renderHTMLTemplate(w, 200, statsHTML, stats)
}

func parseUnixTimeStamp(s string) time.Time {
	ts, err := strconv.Atoi(s)
	if err != nil {
		return time.Unix(0, 0)
	}

	return time.Unix(int64(ts), 0)
}

func writeAPIError(w http.ResponseWriter, httpStatus int, e error) {
	data := struct {
		Error string `json:"error"`
	}{
		e.Error(),
	}
	renderJSONResponse(w, httpStatus, data)
}

func writeAPIIcons(w http.ResponseWriter, url string, icons []besticon.Icon) {
	data := struct {
		URL   string          `json:"url"`
		Icons []besticon.Icon `json:"icons"`
	}{
		url,
		icons,
	}
	renderJSONResponse(w, 200, data)
}

func renderJSONResponse(w http.ResponseWriter, httpStatus int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	enc := json.NewEncoder(w)
	enc.Encode(data)
}

type pageInfo struct {
	URL   string
	Icons []besticon.Icon
	Error error
}

type stats struct {
	Stats []pair
}
type pair struct {
	Name, Value string
}

func (pi pageInfo) Host() string {
	u := pi.URL
	url, _ := url.Parse(u)
	if url != nil && url.Host != "" {
		return url.Host
	}
	return pi.URL
}

func (pi pageInfo) Best() string {
	if len(pi.Icons) > 0 {
		best := pi.Icons[0]
		return best.URL
	}
	return ""
}

func renderHTMLTemplate(w http.ResponseWriter, httpStatus int, templ *template.Template, data interface{}) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(httpStatus)

	err := templ.Execute(w, data)
	if err != nil {
		err = fmt.Errorf("server: could not generate output: %s", err)
		logger.Print(err)
		w.Write([]byte(err.Error()))
	}
}

func startServer(port int) {
	serveAsset("/", "besticon/iconserver/assets/index.html")

	http.HandleFunc("/icons", iconsHandler)
	http.HandleFunc("/api/icons", apiHandler)
	http.HandleFunc("/stats", statsHandler)

	serveAsset("/pure-0.5.0-min.css", "besticon/iconserver/assets/pure-0.5.0-min.css")
	serveAsset("/grids-responsive-0.5.0-min.css", "besticon/iconserver/assets/grids-responsive-0.5.0-min.css")
	serveAsset("/main-min.css", "besticon/iconserver/assets/main-min.css")

	serveAsset("/icon.svg", "besticon/iconserver/assets/icon.svg")
	serveAsset("/favicon.ico", "besticon/iconserver/assets/favicon.ico")
	serveAsset("/apple-touch-icon.png", "besticon/iconserver/assets/apple-touch-icon.png")

	logger.Print("Starting server on port ", port, "...")
	e := http.ListenAndServe(":"+strconv.Itoa(port), NewLoggingMux())
	if e != nil {
		fmt.Printf("cannot start server: %s\n", e)
	}
}

func serveAsset(path string, assetPath string) {
	http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		assetInfo, err := assets.AssetInfo(assetPath)
		if err != nil {
			panic(err)
		}

		http.ServeContent(w, r, assetInfo.Name(), assetInfo.ModTime(), bytes.NewReader(assets.MustAsset(assetPath)))
	})
}

func main() {
	port := flag.Int("port", 0, "Port in server mode")
	flag.Parse()

	if *port > 0 {
		startServer(*port)
	} else {
		flag.PrintDefaults()
	}
}

func init() {
	iconsHTML = templateFromAsset("besticon/iconserver/assets/icons.html", "icons.html")
	statsHTML = templateFromAsset("besticon/iconserver/assets/stats.html", "stats.html")
}

func templateFromAsset(assetPath, templateName string) *template.Template {
	bytes := assets.MustAsset(assetPath)
	return template.Must(template.New(templateName).Parse(string(bytes)))
}

var iconsHTML *template.Template
var statsHTML *template.Template

var logger = log.New(os.Stdout, "besticon: ", log.LstdFlags|log.Ltime)

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

func NewLoggingMux() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		writer := loggingWriter{w, 0, 0}
		http.DefaultServeMux.ServeHTTP(&writer, req)
		end := time.Now()
		duration := end.Sub(start)

		logger.Printf("%s %s %d \"%s\" %s %v %d",
			req.Method,
			req.URL,
			writer.status,
			req.UserAgent(),
			req.Referer(),
			duration,
			writer.length,
		)
	}
}

var (
	fetchCount  = expvar.NewInt("fetchCount")
	fetchErrors = expvar.NewInt("fetchErrors")
)
