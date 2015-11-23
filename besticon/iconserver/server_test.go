package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mat/besticon/besticon"
)

func TestGetIndex(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	indexHandler(w, req)

	assertStringEquals(t, "200", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	assertStringContains(t, w.Body.String(), "<title>The Favicon Finder</title>")
}

func TestGetIcons(t *testing.T) {
	req, err := http.NewRequest("GET", "/icons?url=apple.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	iconsHandler(w, req)

	assertStringEquals(t, "200", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	assertStringContains(t, w.Body.String(), "Icons on apple.com")

	assertStringContains(t, w.Body.String(), "<img src='http://www.apple.com/favicon.ico'")
	assertStringContains(t, w.Body.String(), "<a href='http://www.apple.com/favicon.ico'>")
	assertStringContains(t, w.Body.String(), "<td class='dimensions'>32x32</td>")
}

func TestGetAllIcons(t *testing.T) {
	req, err := http.NewRequest("GET", "/allicons.json?url=apple.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	alliconsHandler(w, req)

	assertStringEquals(t, "200", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "application/json", w.Header().Get("Content-Type"))

	assertStringContains(t, w.Body.String(), `"url":"http://www.apple.com/favicon.ico"`)
	assertStringContains(t, w.Body.String(), `"width":32`)
	assertStringContains(t, w.Body.String(), `"height":32`)
}

func TestGetApiIconsWithMaxAge(t *testing.T) {
	req, err := http.NewRequest("GET", "/allicons.json?url=apple.com&max_age=2h", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	alliconsHandler(w, req)

	assertStringEquals(t, "200", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "application/json", w.Header().Get("Content-Type"))

	assertStringContains(t, w.Body.String(), `"url":"http://www.apple.com/favicon.ico"`)
	assertStringContains(t, w.Body.String(), `"width":32`)
	assertStringContains(t, w.Body.String(), `"height":32`)
}

func TestGetObsoleteApiRedirect(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/icons?url=http%3A%2F%2Fapple.com&i_am_feeling_lucky=yes", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	obsoleteAPIHandler(w, req)

	assertStringEquals(t, "302", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "/icon?size=120&url=http%3A%2F%2Fapple.com&i_am_feeling_lucky=yes", w.Header().Get("Location"))
}

func TestGet404(t *testing.T) {
	req, err := http.NewRequest("GET", "/does-not-exist", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	indexHandler(w, req)

	assertStringEquals(t, "404", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	assertStringContains(t, w.Body.String(), "The requested page does not exist :-(")
}

func assertStringContains(t *testing.T, haystack string, needle string) {
	if !strings.Contains(haystack, needle) {
		fail(t, fmt.Sprintf("Expected '%s' to be contained in '%s'", needle, haystack))
	}
}

func assertStringEquals(t *testing.T, expected string, actual string) {
	if expected != actual {
		fail(t, fmt.Sprintf("Expected '%s' to be '%s'", actual, expected))
	}
}

func fail(t *testing.T, failureMessage string) {
	t.Errorf("\t%s\n"+
		"\r\t",
		failureMessage)
}

func init() {
	besticon.SetLogOutput(ioutil.Discard)
}
