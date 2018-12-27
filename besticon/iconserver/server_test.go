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
	assertStringEquals(t, "max-age=2592000", w.Header().Get("Cache-Control"))
	assertStringEquals(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	assertStringContains(t, w.Body.String(), "Icons on apple.com")

	assertStringContains(t, w.Body.String(), "<img src='https://www.apple.com/favicon.ico'")
	assertStringContains(t, w.Body.String(), "<a href='https://www.apple.com/favicon.ico'>")
	assertStringContains(t, w.Body.String(), "<td class='dimensions'>64x64</td>")
}

func TestGetIcon(t *testing.T) {
	req, err := http.NewRequest("GET", "/icon?url=apple.com&size=120", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	iconHandler(w, req)

	assertStringEquals(t, "302", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "max-age=2592000", w.Header().Get("Cache-Control"))
	assertStringEquals(t, "https://www.apple.com/apple-touch-icon.png", w.Header().Get("Location"))
}

func TestGetIconWithFallBackURL(t *testing.T) {
	req, err := http.NewRequest("GET", "/icon?url=apple.com&size=400&fallback_icon_url=http%3A%2F%2Fexample.com", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	iconHandler(w, req)

	assertStringEquals(t, "302", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "max-age=2592000", w.Header().Get("Cache-Control"))
	assertStringEquals(t, "http://example.com", w.Header().Get("Location"))
}

func TestGetIconWith404Page(t *testing.T) {
	req, err := http.NewRequest("GET", "/icons?size=32&url=httpbin.org/status/404", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	iconHandler(w, req)

	assertStringEquals(t, "302", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "/lettericons/H-32.png", w.Header().Get("Location"))
}

func TestGet404IconWithFallbackColor(t *testing.T) {
	req, err := http.NewRequest("GET", "/icons?size=32&url=httpbin.org/status/404&fallback_icon_color=123456", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	iconHandler(w, req)

	assertStringEquals(t, "302", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "/lettericons/H-32-123456.png", w.Header().Get("Location"))
}

func TestGet404IconWithInvalidFallbackColor(t *testing.T) {
	req, err := http.NewRequest("GET", "/icons?size=32&url=httpbin.org/status/404&fallback_icon_color=zz", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	iconHandler(w, req)

	assertStringEquals(t, "302", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "/lettericons/H-32.png", w.Header().Get("Location"))
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
	assertStringEquals(t, "max-age=2592000", w.Header().Get("Cache-Control"))

	assertStringContains(t, w.Body.String(), `"url":"https://www.apple.com/favicon.ico"`)
	assertStringContains(t, w.Body.String(), `"width":64`)
	assertStringContains(t, w.Body.String(), `"height":64`)

	// Make sure we don't return inlined image data
	assertDoesNotExceed(t, len(w.Body.String()), 2000)
}

func TestGetPopular(t *testing.T) {
	req, err := http.NewRequest("GET", "/popular", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	popularHandler(w, req)

	assertStringEquals(t, "200", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "text/html; charset=utf-8", w.Header().Get("Content-Type"))

	assertStringContains(t, w.Body.String(), `Icon Examples`)
	assertStringContains(t, w.Body.String(), `github.com`)
}

func TestGetLetterIcon(t *testing.T) {
	req, err := http.NewRequest("GET", "/lettericons/M-144-EFC25D.png", nil)
	if err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	lettericonHandler(w, req)

	assertStringEquals(t, "200", fmt.Sprintf("%d", w.Code))
	assertStringEquals(t, "image/png", w.Header().Get("Content-Type"))
	assertStringEquals(t, "max-age=31536000", w.Header().Get("Cache-Control"))
	assertIntegerInInterval(t, 1500, 1800, w.Body.Len())
}

func TestGetBadLetterIconPath(t *testing.T) {
	req, err := http.NewRequest("GET", "/lettericons/--120.png", nil)
	if err != nil {
		log.Fatal(err)
	}

	w := httptest.NewRecorder()
	lettericonHandler(w, req)

	assertStringEquals(t, "400", fmt.Sprintf("%d", w.Code))
	assertStringContains(t, w.Body.String(), `wrong format for lettericons/ path`)
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

func assertIntegerEquals(t *testing.T, expected int, actual int) {
	if expected != actual {
		fail(t, fmt.Sprintf("Expected %d to be %d", actual, expected))
	}
}

func assertIntegerInInterval(t *testing.T, lower int, upper int, actual int) {
	if actual < lower || actual > upper {
		fail(t, fmt.Sprintf("Expected %d to be in interval [%d,%d]", actual, lower, upper))
	}
}

func assertDoesNotExceed(t *testing.T, actual int, maximum int) {
	if actual >= maximum {
		fail(t, fmt.Sprintf("Expected '%d' to be < '%d'", actual, maximum))
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
