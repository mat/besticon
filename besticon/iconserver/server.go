package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"

	"github.com/mat/besticon/besticon"
)

func iconsHandler(w http.ResponseWriter, r *http.Request) {
	lg(r)

	url := r.FormValue(urlParam)
	if len(url) == 0 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	icons, e := besticon.FetchIcons(url)
	switch {
	case e != nil:
		writePageWithError(w, pageInfo{URL: url, Error: e})
	case len(icons) <= 0:
		errNoIcons := errors.New("this poor site has no icons at all :-(")
		writePageWithError(w, pageInfo{URL: url, Error: errNoIcons})
	default:
		writePage(w, pageInfo{Icons: icons, URL: url})
	}
}

const urlParam = "url"
const bestParam = "i_am_feeling_lucky"

func apiHandler(w http.ResponseWriter, r *http.Request) {
	lg(r)

	url := r.FormValue(urlParam)
	if len(url) == 0 {
		errMissingURL := errors.New("need url query parameter")
		writeAPIError(w, http.StatusBadRequest, errMissingURL)
		return
	}

	bestMode := r.FormValue(bestParam) == "yes"
	if bestMode {
		icon, e := besticon.FetchBestIcon(url)
		if e != nil {
			writeAPIError(w, http.StatusNotFound, e)
			return
		}

		http.Redirect(w, r, icon.URL, http.StatusFound)
	} else {
		icons, e := besticon.FetchIcons(url)
		if e != nil {
			writeAPIError(w, http.StatusNotFound, e)
			return
		}

		writeAPIIcons(w, url, icons)
	}
}

func writeAPIError(w http.ResponseWriter, status int, e error) {
	data := struct {
		Error string `json:"error"`
	}{
		e.Error(),
	}
	writeJSONResponse(w, status, data)
}

func writeAPIIcons(w http.ResponseWriter, url string, icons []besticon.Icon) {
	data := struct {
		URL   string          `json:"url"`
		Icons []besticon.Icon `json:"icons"`
	}{
		url,
		icons,
	}
	writeJSONResponse(w, 200, data)
}

func writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
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

func writePageWithError(w http.ResponseWriter, pi pageInfo) {
	w.WriteHeader(http.StatusNotFound)
	writePage(w, pi)
}

func writePage(w http.ResponseWriter, pi pageInfo) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")

	e := iconsHTML.Execute(w, pi)
	if e != nil {
		e = fmt.Errorf("server: could not generate output: %s", e)
		logger.Print(e)
		w.Write([]byte(e.Error()))
	}
}

func startServer(port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(root, "/assets/index.html"))
	})
	http.HandleFunc("/icons", iconsHandler)
	http.HandleFunc("/api/icons", apiHandler)

	http.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(root, r.URL.Path))
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(root, "/assets/favicon.ico"))
	})
	http.HandleFunc("/apple-touch-icon.png", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path.Join(root, "/assets/apple-touch-icon.png"))
	})

	e := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if e != nil {
		fmt.Printf("cannot start server: %s\n", e)
	}
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
	root = os.Getenv("ICON_SERVER_ROOT")
	if root == "" {
		root = "./besticon/iconserver"
	}
	iconsHTML = template.Must(template.ParseFiles(path.Join(root, "/assets/icons.html")))
}

var root string
var iconsHTML *template.Template

var logger = log.New(os.Stdout, "besticon: ", log.LstdFlags|log.Lmicroseconds)

func lg(r *http.Request) {
	bytes, _ := json.Marshal(r)
	logger.Print(string(bytes))
}
