package besticon

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/mat/besticon/vcr"
)

//
// Big list of tests for IconFinder.FetchIcons. Responses are cached on disk
// using our VCR file format. No network connectivity is required unless you are
// adding more tests.
//

type testFetch struct {
	url   string
	icons []testFetchIcon
}

type testFetchIcon struct {
	url    string
	width  int
	format string
}

func TestFetchIcons(t *testing.T) {
	tests := []testFetch{
		// alibaba - base tag without scheme
		{"http://alibaba.com", []testFetchIcon{
			{"http://is.alicdn.com/simg/single/icon/favicon.ico", 16, "ico"},
			{"http://www.alibaba.com/favicon.ico", 16, "ico"},
		}},

		// aol - has one pixel gifs
		{"http://aol.com", []testFetchIcon{
			{"http://www.aol.com/favicon.ico", 32, "ico"},
			{"http://www.aol.com/favicon.ico?v=2", 32, "ico"},
		}},

		// archive.org - has jpg
		{"https://archive.org", []testFetchIcon{
			{"https://archive.org/apple-touch-icon-precomposed.png", 180, "png"},
			{"https://archive.org/apple-touch-icon.png", 180, "png"},
			{"https://archive.org/images/glogo.jpg", 40, "jpg"},
			{"https://archive.org/favicon.ico", 32, "ico"},
		}},

		// ard - should sort by size
		{"http://ard.de", []testFetchIcon{
			{"http://www.ard.de/ARD-144.png", 144, "png"},
			{"http://www.ard.de/apple-touch-icon-precomposed.png", 144, "png"},
			{"http://www.ard.de/apple-touch-icon.png", 144, "png"},
			{"http://www.ard.de/favicon.ico", 144, "ico"},
		}},

		// aws.amazon.com - this one has a base url
		{"http://aws.amazon.com", []testFetchIcon{
			{"http://a0.awsstatic.com/main/images/site/touch-icon-ipad-144-precomposed.png", 144, "png"},
			{"http://a0.awsstatic.com/main/images/site/touch-icon-iphone-114-precomposed.png", 114, "png"},
			{"http://a0.awsstatic.com/main/images/site/favicon.ico", 16, "ico"},
			{"http://aws.amazon.com/favicon.ico", 16, "ico"},
		}},

		// car2go - relative urls
		{"http://car2go.com", []testFetchIcon{
			{"https://www.car2go.com/media/assets/patterns/static/img/favicon.ico", 16, "ico"},
		}},

		// daringfireball
		{"http://daringfireball.net", []testFetchIcon{
			{"http://daringfireball.net/graphics/apple-touch-icon.png", 314, "png"},
			{"http://daringfireball.net/favicon.ico", 32, "ico"},
			{"http://daringfireball.net/graphics/favicon.ico?v=005", 32, "ico"},
		}},

		// dnevnik - capitalized icon tag
		{"http://www.dnevnik.bg", []testFetchIcon{
			{"http://www.dnevnik.bg/images/layout/apple-touch-icon.png", 180, "png"},
			{"http://www.dnevnik.bg/apple-touch-icon.png", 129, "png"},
			{"http://www.dnevnik.bg/favicon.ico", 32, "ico"},
			{"http://www.dnevnik.bg/images/layout/favicon.ico", 16, "ico"},
		}},

		// eat24 - has base tag
		{"http://eat24.com", []testFetchIcon{
			// later - for svg
			// {"http://eat24hours.com/static/v4/images/favicon.svg", 9999, "svg" },
			{"http://eat24hours.com/favicon.ico", 16, "ico"},
		}},

		// example.com - no icons
		{"http://example.com", []testFetchIcon{}},

		// github
		{"http://github.com", []testFetchIcon{
			// later - for svg
			// {"https://assets-cdn.github.com/pinned-octocat.svg", 9999, "svg" },
			{"https://github.com/apple-touch-icon-144.png", 144, "png"},
			{"https://github.com/apple-touch-icon.png", 120, "png"},
			{"https://github.com/apple-touch-icon-114.png", 114, "png"},
			{"https://github.com/apple-touch-icon-precomposed.png", 57, "png"},
			{"https://assets-cdn.github.com/favicon.ico", 32, "ico"},
			{"https://github.com/favicon.ico", 32, "ico"},
		}},

		// kicktipp.de
		{"http://kicktipp.de", []testFetchIcon{
			{"http://info.kicktipp.de/assets/img/jar_cb333387130/assets/img/logos/apple-touch-icon-57x57-precomposed.png", 57, "png"},
			{"http://www.kicktipp.de/apple-touch-icon-precomposed.png", 57, "png"},
			{"http://www.kicktipp.de/apple-touch-icon.png", 57, "png"},
			{"http://www.kicktipp.de/favicon.ico", 32, "gif"},
			{"http://info.kicktipp.de/assets/img/jar_cb1652512069/assets/img/logos/favicon.png", 16, "png"},
		}},

		// netflix - has cookie redirects
		{"http://netflix.com", []testFetchIcon{
			{"https://assets.nflxext.com/us/ffe/siteui/common/icons/nficon2016.png", 64, "png"},
			{"https://assets.nflxext.com/us/ffe/siteui/common/icons/nficon2016.ico", 64, "ico"},
			{"https://www.netflix.com/favicon.ico", 64, "ico"},
		}},

		// storage.googleapis - has bad http response
		{"https://storage.googleapis.com", []testFetchIcon{
			{"https://storage.googleapis.com/favicon.ico", 32, "png"},
		}},
	}

	for _, test := range tests {
		fmt.Println("===========================================")
		fmt.Printf("= TestFetchIcons %s \n", test.url)
		fmt.Println("===========================================")

		// no errors expected
		actualIcons, _, err := fetchIconsWithVCR(test.url)
		assertEquals(t, nil, err)

		// now compare icons
		assertEquals(t, len(test.icons), len(actualIcons))
		for i := range test.icons {
			assertEquals(t, test.icons[i].url, actualIcons[i].URL)
			assertEquals(t, test.icons[i].width, actualIcons[i].Width)
			assertEquals(t, test.icons[i].format, actualIcons[i].Format)
		}
	}
}

//
// This is our list of tests for IconFinder.IconInSizeRange.
//

type testIconInSizeRange struct {
	url       string
	sizeRange SizeRange
	winner    string
}

func TestIconInSizeRange(t *testing.T) {
	tests := []testIconInSizeRange{
		{"http://car2go.com", SizeRange{80, 120, 200}, ""},
		{"http://daringfireball.net", SizeRange{20, 80, 500}, "http://daringfireball.net/graphics/apple-touch-icon.png"},
		{"http://eat24.com", SizeRange{120, 150, 500}, ""},
		{"http://kicktipp.de", SizeRange{20, 80, 500}, "http://info.kicktipp.de/assets/img/jar_cb333387130/assets/img/logos/apple-touch-icon-57x57-precomposed.png"},

		// https://github.com/mat/besticon/issues/28
		{"https://random.org", SizeRange{16, 32, 64}, "https://www.random.org/favicon.ico"},

		// This test can only work because with HostOnlyDomains accordingly
		{"http://youtube.com/does-not-exist", SizeRange{0, 80, 200}, "https://s.ytimg.com/yts/img/favicon_96-vfldSA3ca.png"},
	}

	for _, test := range tests {
		fmt.Println("===========================================")
		fmt.Printf("= TestIconInSizeRange %s \n", test.url)
		fmt.Println("===========================================")

		// no errors expected
		_, finder, err := fetchIconsWithVCR(test.url)
		assertEquals(t, nil, err)

		// now compare icons
		actualIcon := finder.IconInSizeRange(test.sizeRange)
		if actualIcon == nil {
			assertEquals(t, test.winner, "")
		} else {
			assertEquals(t, test.winner, actualIcon.URL)
		}
	}
}

//
// other tests
//

func TestMortenmøllerWithIDNAHost(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("https://mortenmøller.dk")
	assertEquals(t, nil, err)
	assertEquals(t, 13, len(actualImages))
}

func TestMainColorForIconsWithBrokenImageData(t *testing.T) {
	icn := Icon{Format: "png", ImageData: []byte("broken-image-data")}
	colr := MainColorForIcons([]Icon{icn})
	assertEquals(t, (*color.RGBA)(nil), colr)
}

func TestImageSizeDetection(t *testing.T) {
	assertEquals(t, 1, getImageWidthForFile("testdata/pixel.gif"))
	assertEquals(t, 1, getImageWidthForFile("testdata/pixel.jpg"))
	assertEquals(t, 1, getImageWidthForFile("testdata/pixel.png"))
	assertEquals(t, 48, getImageWidthForFile("testdata/favicon.ico"))
}

func TestParseSizeRange(t *testing.T) {
	// This single num behaviour ensures backwards compatibility for
	// people who pant (at least) pixel perfect icons.
	sizeRange, err := ParseSizeRange("120")
	check(err)
	assertEquals(t, &SizeRange{120, 120, MaxIconSize}, sizeRange)

	sizeRange, err = ParseSizeRange("0..120..256")
	check(err)
	assertEquals(t, &SizeRange{0, 120, 256}, sizeRange)

	sizeRange, err = ParseSizeRange("120..120..120")
	check(err)
	assertEquals(t, &SizeRange{120, 120, 120}, sizeRange)

	_, err = ParseSizeRange("")
	assertEquals(t, errBadSize, err)

	_, err = ParseSizeRange(" ")
	assertEquals(t, errBadSize, err)

	// Max < Perfect not allowed
	_, err = ParseSizeRange("16..120..80")
	assertEquals(t, errBadSize, err)

	// Perfect < Min  not allowed
	_, err = ParseSizeRange("120..16..80")
	assertEquals(t, errBadSize, err)

	// Min too small
	_, err = ParseSizeRange("-1..2..3")
	assertEquals(t, errBadSize, err)

	// Max too big
	_, err = ParseSizeRange("1..2..501")
	assertEquals(t, errBadSize, err)
}

func TestGetenvOrFallback(t *testing.T) {
	os.Setenv("MY_ENV", "some-value")
	assertEquals(t, "some-value", getenvOrFallback("MY_ENV", "fallback-should-NOT-be-used"))

	os.Setenv("MY_ENV", "")
	assertEquals(t, "fallback-should-be-used", getenvOrFallback("MY_ENV", "fallback-should-be-used"))

	assertEquals(t, "fallback-should-be-used", getenvOrFallback("key-does-not-exist", "fallback-should-be-used"))
}

func TestParseSize(t *testing.T) {
	size, ok := parseSize("120")
	assertEquals(t, ok, true)
	assertEquals(t, 120, size)

	_, ok = parseSize("")
	assertEquals(t, ok, false)

	_, ok = parseSize("-10")
	assertEquals(t, ok, false)
}

func TestAbsoluteURL(t *testing.T) {
	baseURL, e := url.Parse("http://car2go.com")
	check(e)
	u, e := absoluteURL(baseURL, "/../../media/favicon.ico")
	check(e)
	assertEquals(t, "http://car2go.com/media/favicon.ico", u)
}

func TestIsSVG(t *testing.T) {
	invalid := [][]byte{
		[]byte(""),
		[]byte("<html></html>"),
		mustReadFile("testdata/favicon.ico"),
	}
	for _, data := range invalid {
		assertEquals(t, isSVG(data), false)
	}

	valid := [][]byte{
		[]byte("<svg></svg>"),
		[]byte("<!-- comment --><svg></svg>"),
		[]byte("<?xml?><!DOCTYPE svg><svg></svg>"),
		mustReadFile("testdata/svg.svg"),
	}
	for _, data := range valid {
		assertEquals(t, isSVG(data), true)
	}
}

//
// helpers
//

const testdataDir = "testdata/"

func fetchIconsWithVCR(s string) ([]Icon, *IconFinder, error) {
	URL, _ := url.Parse(s)
	path := fmt.Sprintf("%s%s.vcr", testdataDir, URL.Host)

	// build client
	client, f, err := vcr.Client(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	setHTTPClient(client)

	// fetch
	finder := IconFinder{}
	finder.HostOnlyDomains = []string{"youtube.com"}
	icons, err := finder.FetchIcons(s)
	return icons, &finder, err
}

func getImageWidthForFile(filename string) int {
	f, err := os.Open(filename)
	check(err)
	defer f.Close()

	icfg, _, err := image.DecodeConfig(f)
	check(err)
	return icfg.Width
}

func mustReadFile(filename string) []byte {
	bytes, e := ioutil.ReadFile(filename)
	check(e)
	return bytes
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func assertEquals(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		fail(t, fmt.Sprintf("Not equal: %#v (expected)\n"+
			"        != %#v (actual)", expected, actual))
	}
}

func fail(t *testing.T, failureMessage string) {
	t.Errorf("\t%s\n"+
		"\r\t",
		failureMessage)
}

func init() {
	keepImageBytes = false
}
