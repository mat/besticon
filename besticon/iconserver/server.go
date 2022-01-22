package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mat/besticon/besticon"
	"github.com/mat/besticon/besticon/iconserver/assets"
	"github.com/mat/besticon/lettericon"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"
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

	finder := newIconFinder()

	formats := r.FormValue("formats")
	if formats != "" {
		finder.FormatsAllowed = strings.Split(r.FormValue("formats"), ",")
	}

	icons, e := finder.FetchIcons(url)
	switch {
	case e != nil:
		renderHTMLTemplate(w, 404, iconsHTML, pageInfo{URL: url, Error: e})
	case len(icons) == 0:
		errNoIcons := errors.New("this poor site has no icons at all :-(")
		renderHTMLTemplate(w, 404, iconsHTML, pageInfo{URL: url, Error: errNoIcons})
	default:
		addCacheControl(w, cacheDurationSeconds)
		renderHTMLTemplate(w, 200, iconsHTML, pageInfo{Icons: icons, URL: url})
	}
}

func iconHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if len(url) == 0 {
		writeAPIError(w, 400, errors.New("need url parameter"))
		return
	}

	sizeRange, err := besticon.ParseSizeRange(r.FormValue("size"))
	if err != nil {
		writeAPIError(w, 400, errors.New("bad size parameter"))
		return
	}

	finder := newIconFinder()
	formats := r.FormValue("formats")
	if formats != "" {
		finder.FormatsAllowed = strings.Split(r.FormValue("formats"), ",")
	}

	finder.FetchIcons(url)

	icon := finder.IconInSizeRange(*sizeRange)
	if icon != nil {
		returnIcon(w, r, icon.URL)
		return
	}

	fallbackIconURL := r.FormValue("fallback_icon_url")
	if fallbackIconURL != "" {
		returnIcon(w, r, fallbackIconURL)
		return
	}

	iconColor := finder.MainColorForIcons()
	letter := lettericon.MainLetterFromURL(url)

	fallbackColorHex := r.FormValue("fallback_icon_color")
	if iconColor == nil && fallbackColorHex != "" {
		color, err := lettericon.ColorFromHex(fallbackColorHex)
		if err == nil {
			iconColor = color
		}
	}

	// We support both PNG and SVG fallback. Only return SVG if requested.
	format := "png"
	if includesString(finder.FormatsAllowed, "svg") {
		format = "svg"
	}
	redirectPath := lettericon.IconPath(letter, fmt.Sprintf("%d", sizeRange.Perfect), iconColor, format)
	redirectWithCacheControl(w, r, redirectPath)
}

func popularHandler(w http.ResponseWriter, r *http.Request) {
	iconSize, err := strconv.Atoi(r.FormValue("iconsize"))
	if iconSize > besticon.MaxIconSize || iconSize < besticon.MinIconSize || err != nil {
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
	urlParam = "url"
)

func alliconsHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue(urlParam)
	if len(url) == 0 {
		errMissingURL := errors.New("need url query parameter")
		writeAPIError(w, 400, errMissingURL)
		return
	}

	finder := newIconFinder()
	formats := r.FormValue("formats")
	if formats != "" {
		finder.FormatsAllowed = strings.Split(r.FormValue("formats"), ",")
	}

	icons, e := finder.FetchIcons(url)
	if e != nil {
		writeAPIError(w, 404, e)
		return
	}

	addCacheControl(w, cacheDurationSeconds)
	writeAPIIcons(w, url, icons)
}

func lettericonHandler(w http.ResponseWriter, r *http.Request) {
	charParam, col, size, format := lettericon.ParseIconPath(r.URL.Path)
	if charParam == "" || col == nil || size <= 0 || format == "" {
		writeAPIError(w, 400, errors.New("wrong format for lettericons/ path, must look like lettericons/M-144-EFC25D.png or M-EFC25D.svg"))
		return
	}

	if format == "svg" {
		w.Header().Add(contentType, imageSVG)
		lettericon.RenderSVG(charParam, col, w)
	} else {
		w.Header().Add(contentType, imagePNG)
		lettericon.RenderPNG(charParam, col, size, w)
	}
	addCacheControl(w, oneYear)
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
	renderJSONResponse(w, 200, data)
}

const (
	contentType     = "Content-Type"
	applicationJSON = "application/json"
	imagePNG        = "image/png"
	imageSVG        = "image/svg+xml"
)

func renderJSONResponse(w http.ResponseWriter, httpStatus int, data interface{}) {
	w.Header().Add(contentType, applicationJSON)
	w.WriteHeader(httpStatus)
	enc := json.NewEncoder(w)
	enc.Encode(data)
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

func startServer(port string, address string) {
	registerHandler("/", indexHandler)
	registerHandler("/icons", iconsHandler)
	registerHandler("/icon", iconHandler)
	registerHandler("/popular", popularHandler)
	registerHandler("/allicons.json", alliconsHandler)
	registerHandler("/lettericons/", lettericonHandler)

	serveAsset("/pure-0.5.0-min.css", "pure-0.5.0-min.css", oneYear)
	serveAsset("/grids-responsive-0.5.0-min.css", "grids-responsive-0.5.0-min.css", oneYear)
	serveAsset("/main-min.css", "main-min.css", oneYear)

	serveAsset("/icon.svg", "icon.svg", oneYear)
	serveAsset("/favicon.ico", "favicon.ico", oneYear)
	serveAsset("/apple-touch-icon.png", "apple-touch-icon.png", oneYear)

	metricsPath := getenvOrFallback("METRICS_PATH", "/metrics")

	if metricsPath != "disable" {
		if !strings.HasPrefix(metricsPath, "/") {
			logger.Fatalf("METRICS_PATH must start with a slash")
		}

		http.Handle(metricsPath, promhttp.Handler())
	}

	addr := address + ":" + port
	logger.Print("Starting server on ", addr, "...")
	e := http.ListenAndServe(addr, httpHandler())
	if e != nil {
		logger.Fatalf("cannot start server: %s\n", e)
	}
}

func httpHandler() http.Handler {
	corsEnabled := getTrueFromEnv("CORS_ENABLED")
	if corsEnabled {
		logger.Print("Enabling CORS middleware")
		return corsHandler(newLoggingMux())
	} else {
		return newLoggingMux()
	}
}

func corsHandler(mux http.HandlerFunc) http.Handler {
	corsOpts := cors.Options{
		AllowedOrigins:   stringSliceFromEnv("CORS_ALLOWED_ORIGINS"),
		AllowedMethods:   stringSliceFromEnv("CORS_ALLOWED_METHODS"),
		AllowedHeaders:   stringSliceFromEnv("CORS_ALLOWED_HEADERS"),
		AllowCredentials: getTrueFromEnv("CORS_ALLOW_CREDENTIALS"),
		Debug:            getTrueFromEnv("CORS_DEBUG"),
	}
	return cors.New(corsOpts).Handler(mux)
}

const (
	cacheControl = "Cache-Control"
	oneYear      = 365 * 24 * 3600
)

func returnIcon(w http.ResponseWriter, r *http.Request, iconURL string) {
	if os.Getenv("SERVER_MODE") == "download" {
		downloadAndReturn(w, r, iconURL)
	} else {
		redirectWithCacheControl(w, r, iconURL)
	}
}

func downloadAndReturn(w http.ResponseWriter, r *http.Request, iconURL string) {
	response, e := besticon.Get(iconURL)
	if e != nil {
		redirectWithCacheControl(w, r, iconURL)
	}

	b, e := besticon.GetBodyBytes(response)
	if e != nil {
		redirectWithCacheControl(w, r, iconURL)
	}

	addCacheControl(w, cacheDurationSeconds)
	w.Write(b)
}

func redirectWithCacheControl(w http.ResponseWriter, r *http.Request, redirectURL string) {
	addCacheControl(w, cacheDurationSeconds)
	http.Redirect(w, r, redirectURL, 302)
}

func addCacheControl(w http.ResponseWriter, maxAge int) {
	w.Header().Add(cacheControl, fmt.Sprintf("max-age=%d", maxAge))
}

func serveAsset(path string, assetPath string, maxAgeSeconds int) {
	registerHandler(path, func(w http.ResponseWriter, r *http.Request) {
		f, err := assets.Assets.Open(assetPath)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			panic(err)
		}

		data, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}

		addCacheControl(w, maxAgeSeconds)

		http.ServeContent(w, r, stat.Name(), stat.ModTime(), bytes.NewReader(data))
	})
}

func registerHandler(path string, f http.HandlerFunc) {
	http.Handle(path, newPrometheusHandler(path, f))
}

func main() {
	fmt.Printf("iconserver %s (%s) (%s) - https://github.com/mat/besticon\n", besticon.VersionString, besticon.BuildDate, runtime.Version())
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "0.0.0.0"
	}
	startServer(port, address)
}

func init() {
	indexHTML = templateFromAsset("index.html", "index.html")
	iconsHTML = templateFromAsset("icons.html", "icons.html")
	popularHTML = templateFromAsset("popular.html", "popular.html")
	notFoundHTML = templateFromAsset("not_found.html", "not_found.html")
}

func templateFromAsset(assetPath, templateName string) *template.Template {
	data, err := assets.Assets.ReadFile(assetPath)
	if err != nil {
		panic(err)
	}
	return template.Must(template.New(templateName).Funcs(funcMap).Parse(string(data)))
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

func newIconFinder() *besticon.IconFinder {
	finder := besticon.IconFinder{}
	if len(hostOnlyDomains) > 0 {
		finder.HostOnlyDomains = hostOnlyDomains
	}

	return &finder
}

var hostOnlyDomains []string
var cacheDurationSeconds int

func init() {
	cacheSize := os.Getenv("CACHE_SIZE_MB")
	if cacheSize == "" {
		besticon.SetCacheMaxSize(32)
	} else {
		n, _ := strconv.Atoi(cacheSize)
		besticon.SetCacheMaxSize(int64(n))
	}

	duration, e := time.ParseDuration(getenvOrFallback("HTTP_MAX_AGE_DURATION", "720h"))
	if e != nil {
		panic(e)
	}
	cacheDurationSeconds = (int)(duration.Seconds())

	hostOnlyDomains = strings.Split(os.Getenv("HOST_ONLY_DOMAINS"), ",")
}

func getTrueFromEnv(s string) bool {
	return getenvOrFallback(s, "") == "true"
}

func stringSliceFromEnv(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func getenvOrFallback(key string, fallbackValue string) string {
	value := os.Getenv(key)
	if len(strings.TrimSpace(value)) != 0 {
		return value
	}
	return fallbackValue
}

func includesString(arr []string, str string) bool {
	for _, e := range arr {
		if e == str {
			return true
		}
	}
	return false
}
