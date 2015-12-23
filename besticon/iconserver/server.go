package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"expvar"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/mat/besticon/besticon"
	"github.com/mat/besticon/besticon/iconserver/assets"
	"github.com/mat/besticon/lettericon"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "" || r.URL.Path == "/" {
		renderHTMLTemplate(w, 200, indexHTML, nil)
	} else {
		renderHTMLTemplate(w, 404, notFoundHTML, nil)
	}
}

func iconsHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue(urlParam)
	if len(url) == 0 {
		http.Redirect(w, r, "/", 302)
		return
	}

	finder := besticon.IconFinder{}

	formats := r.FormValue("formats")
	if formats != "" {
		finder.FormatsAllowed = strings.Split(r.FormValue("formats"), ",")
	}

	e, icons := finder.FetchIcons(url)
	switch {
	case e != nil:
		renderHTMLTemplate(w, 404, iconsHTML, pageInfo{URL: url, Error: e})
	case len(icons) == 0:
		errNoIcons := errors.New("this poor site has no icons at all :-(")
		renderHTMLTemplate(w, 404, iconsHTML, pageInfo{URL: url, Error: errNoIcons})
	default:
		renderHTMLTemplate(w, 200, iconsHTML, pageInfo{Icons: icons, URL: url})
	}
}

func iconHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if len(url) == 0 {
		writeAPIError(w, 400, errors.New("need url parameter"), true)
		return
	}

	size := r.FormValue("size")
	if size == "" {
		writeAPIError(w, 400, errors.New("need size parameter"), true)
		return
	}
	minSize, err := strconv.Atoi(size)
	if err != nil || minSize < 0 || minSize > 1000 {
		writeAPIError(w, 400, errors.New("bad size parameter"), true)
		return
	}

	finder := besticon.IconFinder{}
	formats := r.FormValue("formats")
	if formats != "" {
		finder.FormatsAllowed = strings.Split(r.FormValue("formats"), ",")
	}

	err, _ = finder.FetchIcons(url)
	if err != nil {
		writeAPIError(w, 404, err, true)
		return
	}

	icon := finder.IconWithMinSize(minSize)
	if icon != nil {
		http.Redirect(w, r, icon.URL, 302)
		return
	}

	fallbackIconURL := r.FormValue("fallback_icon_url")
	if fallbackIconURL != "" {
		http.Redirect(w, r, fallbackIconURL, 302)
		return
	}

	iconColor := finder.MainColorForIcons()
	letter := lettericon.MainLetterFromURL(url)
	redirectPath := lettericon.IconPath(letter, size, iconColor)
	http.Redirect(w, r, redirectPath, 302)
}

func popularHandler(w http.ResponseWriter, r *http.Request) {
	iconSize, err := strconv.Atoi(r.FormValue("iconsize"))
	if iconSize > 1000 || iconSize < 10 || err != nil {
		iconSize = 120
	}

	pageInfo := struct {
		URLs        []string
		IconSize    int
		DisplaySize int
	}{
		besticon.PopularSites,
		iconSize,
		iconSize / 2,
	}
	renderHTMLTemplate(w, 200, popularHTML, pageInfo)
}

const (
	urlParam    = "url"
	prettyParam = "pretty"
	maxAge      = "max_age"
)

const defaultMaxAge = time.Duration(604800) * time.Second // 7 days

func alliconsHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue(urlParam)
	if len(url) == 0 {
		errMissingURL := errors.New("need url query parameter")
		writeAPIError(w, 400, errMissingURL, true)
		return
	}

	finder := besticon.IconFinder{}
	formats := r.FormValue("formats")
	if formats != "" {
		finder.FormatsAllowed = strings.Split(r.FormValue("formats"), ",")
	}

	e, icons := finder.FetchIcons(url)
	if e != nil {
		writeAPIError(w, 404, e, true)
		return
	}

	pretty, err := strconv.ParseBool(r.FormValue(prettyParam))
	prettyPrint := (err == nil) && pretty

	writeAPIIcons(w, url, icons, prettyPrint)
}

func lettericonHandler(w http.ResponseWriter, r *http.Request) {
	charParam, col, size := lettericon.ParseIconPath(r.URL.Path)
	if charParam != "" {
		w.Header().Add("Content-Type", "image/png")
		lettericon.Render(charParam, col, size, w)
	} else {
		writeAPIError(w, 400, errors.New("wrong format for lettericons/ path, must look like lettericons/M-144-EFC25D.png"), true)
	}
}

func obsoleteAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("i_am_feeling_lucky") == "yes" {
		http.Redirect(w, r, fmt.Sprintf("/icon?size=120&%s", r.URL.RawQuery), 302)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/allicons.json?%s", r.URL.RawQuery), 302)
	}
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

func writeAPIIcons(w http.ResponseWriter, url string, icons []besticon.Icon, pretty bool) {
	// Don't return whole image data
	newIcons := []besticon.Icon{}
	for _, ico := range icons {
		newIcon := ico
		newIcon.ImageData = nil
		newIcons = append(newIcons, newIcon)
	}

	data := &struct {
		URL   string          `json:"url"`
		Icons []besticon.Icon `json:"icons"`
	}{
		url,
		newIcons,
	}

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

func startServer(port string) {
	registerGzipHandler("/", indexHandler)
	registerGzipHandler("/icons", iconsHandler)
	registerHandler("/icon", iconHandler)
	registerGzipHandler("/popular", popularHandler)
	registerGzipHandler("/allicons.json", alliconsHandler)
	registerHandler("/lettericons/", lettericonHandler)
	registerHandler("/api/icons", obsoleteAPIHandler)

	serveAsset("/pure-0.5.0-min.css", "besticon/iconserver/assets/pure-0.5.0-min.css", oneYear)
	serveAsset("/grids-responsive-0.5.0-min.css", "besticon/iconserver/assets/grids-responsive-0.5.0-min.css", oneYear)
	serveAsset("/main-min.css", "besticon/iconserver/assets/main-min.css", oneYear)

	serveAsset("/icon.svg", "besticon/iconserver/assets/icon.svg", oneYear)
	serveAsset("/favicon.ico", "besticon/iconserver/assets/favicon.ico", oneYear)
	serveAsset("/apple-touch-icon.png", "besticon/iconserver/assets/apple-touch-icon.png", oneYear)
	serveAsset("/robots.txt", "besticon/iconserver/assets/robots.txt", nocache)
	serveAsset("/test-lettericons", "besticon/iconserver/assets/test-lettericons.html", nocache)

	addr := "0.0.0.0:" + port
	logger.Print("Starting server on ", addr, "...")
	e := http.ListenAndServe(addr, newLoggingMux())
	if e != nil {
		logger.Fatalf("cannot start server: %s\n", e)
	}
}

const (
	oneYear = 365 * 24 * 3600
	nocache = -1
)

func serveAsset(path string, assetPath string, maxAgeSeconds int) {
	registerGzipHandler(path, func(w http.ResponseWriter, r *http.Request) {
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

func registerHandler(path string, f http.HandlerFunc) {
	http.Handle(path, newExpvarHandler(path, f))
}

func registerGzipHandler(path string, f http.HandlerFunc) {
	http.Handle(path, gziphandler.GzipHandler(newExpvarHandler(path, f)))
}

func main() {
	fmt.Printf("iconserver %s (%s) - http://icons.better-idea.org\n", besticon.VersionString, runtime.Version())
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	startServer(port)
}

func init() {
	indexHTML = templateFromAsset("besticon/iconserver/assets/index.html", "index.html")
	iconsHTML = templateFromAsset("besticon/iconserver/assets/icons.html", "icons.html")
	popularHTML = templateFromAsset("besticon/iconserver/assets/popular.html", "popular.html")
	notFoundHTML = templateFromAsset("besticon/iconserver/assets/not_found.html", "not_found.html")
}

func templateFromAsset(assetPath, templateName string) *template.Template {
	bytes := assets.MustAsset(assetPath)
	return template.Must(template.New(templateName).Funcs(funcMap).Parse(string(bytes)))
}

var indexHTML *template.Template
var iconsHTML *template.Template
var popularHTML *template.Template
var notFoundHTML *template.Template

var funcMap = template.FuncMap{
	"ImgWidth": imgWidth,
}

func imgWidth(i *besticon.Icon) int {
	return i.Width / 2.0
}

func init() {
	besticon.SetCacheMaxSize(128)

	expvar.Publish("cacheBytes", expvar.Func(func() interface{} { return besticon.GetCacheStats().Bytes }))
	expvar.Publish("cacheItems", expvar.Func(func() interface{} { return besticon.GetCacheStats().Items }))
	expvar.Publish("cacheGets", expvar.Func(func() interface{} { return besticon.GetCacheStats().Gets }))
	expvar.Publish("cacheHits", expvar.Func(func() interface{} { return besticon.GetCacheStats().Hits }))
	expvar.Publish("cacheEvictions", expvar.Func(func() interface{} { return besticon.GetCacheStats().Evictions }))
}
