package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"expvar"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/mat/besticon/besticon"
	"github.com/mat/besticon/besticon/iconserver/assets"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	renderHTMLTemplate(w, 200, indexHTML, nil)
}

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

const (
	urlParam          = "url"
	feelingLuckyParam = "i_am_feeling_lucky"
	prettyParam       = "pretty"
	maxAge            = "max_age"
)

const defaultMaxAge = time.Duration(604800) * time.Second // 7 days

func apiHandler(w http.ResponseWriter, r *http.Request) {
	pretty := r.FormValue(prettyParam) == "yes"
	url := r.FormValue(urlParam)
	if len(url) == 0 {
		errMissingURL := errors.New("need url query parameter")
		writeAPIError(w, 400, errMissingURL, pretty)
		return
	}

	bestOnly := r.FormValue(feelingLuckyParam) == "yes"
	if bestOnly {
		icon, e := fetchBestIcon(url)
		if e != nil {
			writeAPIError(w, 404, e, pretty)
			return
		}

		http.Redirect(w, r, icon.URL, 302)
	} else {
		icons, e := fetchIcons(url)
		if e != nil {
			writeAPIError(w, 404, e, pretty)
			return
		}

		maxAge, err := time.ParseDuration(r.FormValue(maxAge))
		if err != nil || maxAge.Seconds() < 1 {
			maxAge = defaultMaxAge
		}
		writeAPIIcons(w, url, icons, maxAge, pretty)
	}
}

func fetchIcons(url string) ([]besticon.Icon, error) {
	fetchCount.Add(1)
	icons, err := besticon.FetchIcons(url, true)
	if err != nil {
		fetchErrors.Add(1)
	}
	return icons, err
}

func fetchBestIcon(url string) (*besticon.Icon, error) {
	fetchCount.Add(1)
	icon, err := besticon.FetchBestIcon(url, true)
	if err != nil {
		fetchErrors.Add(1)
	}
	return icon, err
}

func writeAPIError(w http.ResponseWriter, httpStatus int, e error, pretty bool) {
	data := struct {
		Error string `json:"error"`
	}{
		e.Error(),
	}

	if pretty {
		renderJSONResponsePretty(w, httpStatus, data)
	} else {
		renderJSONResponse(w, httpStatus, data)
	}
}

func writeAPIIcons(w http.ResponseWriter, url string, icons []besticon.Icon, maxAge time.Duration, pretty bool) {
	data := struct {
		URL   string          `json:"url"`
		Icons []besticon.Icon `json:"icons"`
	}{
		url,
		icons,
	}

	w.Header().Add(cacheControl, fmt.Sprintf("max-age=%d", int64(maxAge.Seconds())))
	if pretty {
		renderJSONResponsePretty(w, 200, data)
	} else {
		renderJSONResponse(w, 200, data)
	}
}

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
	cacheControl    = "Cache-Control"
)

func renderJSONResponse(w http.ResponseWriter, httpStatus int, data interface{}) {
	w.Header().Add(contentType, applicationJSON)
	w.WriteHeader(httpStatus)
	enc := json.NewEncoder(w)
	enc.Encode(data)
}

func renderJSONResponsePretty(w http.ResponseWriter, httpStatus int, data interface{}) {
	w.Header().Add(contentType, applicationJSON)
	w.WriteHeader(httpStatus)
	b, _ := json.MarshalIndent(data, "", "  ")
	w.Write(b)
}

type pageInfo struct {
	URL   string
	Icons []besticon.Icon
	Error error
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
	w.Header().Add(contentType, "text/html; charset=utf-8")
	w.WriteHeader(httpStatus)

	err := templ.Execute(w, data)
	if err != nil {
		err = fmt.Errorf("server: could not generate output: %s", err)
		logger.Print(err)
		w.Write([]byte(err.Error()))
	}
}

func startServer(port int) {
	httpHandle("/", indexHandler)
	httpHandle("/icons", iconsHandler)
	httpHandle("/api/icons", apiHandler)

	serveAsset("/pure-0.5.0-min.css", "besticon/iconserver/assets/pure-0.5.0-min.css", oneYear)
	serveAsset("/grids-responsive-0.5.0-min.css", "besticon/iconserver/assets/grids-responsive-0.5.0-min.css", oneYear)
	serveAsset("/main-min.css", "besticon/iconserver/assets/main-min.css", oneYear)

	serveAsset("/icon.svg", "besticon/iconserver/assets/icon.svg", oneYear)
	serveAsset("/favicon.ico", "besticon/iconserver/assets/favicon.ico", oneYear)
	serveAsset("/apple-touch-icon.png", "besticon/iconserver/assets/apple-touch-icon.png", oneYear)
	serveAsset("/robots.txt", "besticon/iconserver/assets/robots.txt", nocache)

	logger.Print("Starting server on port ", port, "...")
	e := http.ListenAndServe(":"+strconv.Itoa(port), newLoggingMux())
	if e != nil {
		fmt.Printf("cannot start server: %s\n", e)
	}
}

const (
	oneYear = 365 * 24 * 3600
	nocache = -1
)

func serveAsset(path string, assetPath string, maxAgeSeconds int) {
	httpHandle(path, func(w http.ResponseWriter, r *http.Request) {
		assetInfo, err := assets.AssetInfo(assetPath)
		if err != nil {
			panic(err)
		}

		if maxAgeSeconds != nocache {
			w.Header().Add(cacheControl, fmt.Sprintf("max-age=%d", maxAgeSeconds))
		}

		http.ServeContent(w, r, assetInfo.Name(), assetInfo.ModTime(),
			bytes.NewReader(assets.MustAsset(assetPath)))
	})
}

func httpHandle(path string, f http.HandlerFunc) {
	http.Handle(path, gziphandler.GzipHandler(http.HandlerFunc(f)))
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
	indexHTML = templateFromAsset("besticon/iconserver/assets/index.html", "index.html")
	iconsHTML = templateFromAsset("besticon/iconserver/assets/icons.html", "icons.html")
}

func templateFromAsset(assetPath, templateName string) *template.Template {
	bytes := assets.MustAsset(assetPath)
	return template.Must(template.New(templateName).Funcs(funcMap).Parse(string(bytes)))
}

var indexHTML *template.Template
var iconsHTML *template.Template

var funcMap = template.FuncMap{
	"GoogleAnalyticsID": googleAnalyticsID,
	"ImgWidth":          imgWidth,
}

func googleAnalyticsID() string {
	return os.Getenv("ICONS_ANALYTICS_ID")
}

func imgWidth(i *besticon.Icon) int {
	return i.Width / 2.0
}

func init() {
	besticon.SetCacheMaxSize(64)

	expvar.Publish("cacheBytes", expvar.Func(func() interface{} { return besticon.GetCacheStats().Bytes }))
	expvar.Publish("cacheItems", expvar.Func(func() interface{} { return besticon.GetCacheStats().Items }))
	expvar.Publish("cacheGets", expvar.Func(func() interface{} { return besticon.GetCacheStats().Gets }))
	expvar.Publish("cacheHits", expvar.Func(func() interface{} { return besticon.GetCacheStats().Hits }))
	expvar.Publish("cacheEvictions", expvar.Func(func() interface{} { return besticon.GetCacheStats().Evictions }))

	ticker := time.NewTicker(time.Minute * 1)
	go func() {
		for range ticker.C {
			logger.Printf("Cache: %+v", besticon.GetCacheStats())
		}
	}()
}
