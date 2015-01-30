package besticon

import (
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/mat/besticon/vcr"
	"image"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestKicktipp(t *testing.T) {
	actualImages, err := fetchIconsWithVCR("kicktipp.vcr", "http://kicktipp.de")
	check(err)
	expectedImages := []Icon{
		{URL: "http://info.kicktipp.de/assets/img/jar_cb333387130/assets/img/logos/apple-touch-icon-57x57-precomposed.png", Width: 57, Height: 57, Format: "png", Bytes: 1535, Sha1sum: "79aae9e0df7d52ed50ac47c1dd4bd16e2ddf8b4a"},
		{URL: "http://www.kicktipp.de/apple-touch-icon-precomposed.png", Width: 57, Height: 57, Format: "png", Bytes: 1622, Sha1sum: "fd4306aefd9ed2b4983697ca56301d8810a03987"},
		{URL: "http://www.kicktipp.de/apple-touch-icon.png", Width: 57, Height: 57, Format: "png", Bytes: 1622, Sha1sum: "fd4306aefd9ed2b4983697ca56301d8810a03987"},
		{URL: "http://www.kicktipp.de/favicon.ico", Width: 32, Height: 32, Format: "gif", Bytes: 35275, Sha1sum: "09297d0ffe17149c3d4d4a3a3a8c7e8c51932d58"},
		{URL: "http://info.kicktipp.de/assets/img/jar_cb1652512069/assets/img/logos/favicon.png", Width: 16, Height: 16, Format: "png", Bytes: 1820, Sha1sum: "04b49fac810828f6723cd763600af23f0edbde03"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestDaringfireball(t *testing.T) {
	actualImages, err := fetchIconsWithVCR("daringfireball.net.vcr", "http://daringfireball.net")
	check(err)

	expectedImages := []Icon{
		{URL: "http://daringfireball.net/graphics/apple-touch-icon.png", Width: 314, Height: 314, Format: "png", Bytes: 8642, Error: error(nil), Sha1sum: "f47cf7cf13ec1a74049d99d9f1565354b5b20317"},
		{URL: "http://daringfireball.net/favicon.ico", Width: 32, Height: 32, Format: "ico", Bytes: 6518, Error: error(nil), Sha1sum: "c066c70aa1dd2b4347d3095592aac28b55e85d04"},
		{URL: "http://daringfireball.net/graphics/favicon.ico?v=005", Width: 32, Height: 32, Format: "ico", Bytes: 6518, Error: error(nil), Sha1sum: "c066c70aa1dd2b4347d3095592aac28b55e85d04"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestAwsAmazonChangingBaseURL(t *testing.T) {
	actualImages, err := fetchIconsWithVCR("aws.amazon.vcr", "aws.amazon.com")
	check(err)
	expectedImages := []Icon{
		{URL: "http://a0.awsstatic.com/main/images/site/touch-icon-ipad-144-precomposed.png", Width: 144, Height: 144, Format: "png", Bytes: 3944, Sha1sum: "225817df40ff11d083c282d08b49a5ed50fd0f2d"},
		{URL: "http://a0.awsstatic.com/main/images/site/touch-icon-iphone-114-precomposed.png", Width: 114, Height: 114, Format: "png", Bytes: 3081, Sha1sum: "58aabb2a99fcb283710fd200c9feed69c015a29e"},
		{URL: "http://a0.awsstatic.com/main/images/site/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 1150, Sha1sum: "a64f3616e3671b835f8b208b7339518d8b386a08"},
		{URL: "http://aws.amazon.com/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 1150, Sha1sum: "a64f3616e3671b835f8b208b7339518d8b386a08"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestNetflixWitCookieRedirects(t *testing.T) {
	actualImages, err := fetchIconsWithVCR("netflix.vcr", " netflix.com	 ")
	check(err)
	expectedImages := []Icon{
		{URL: "https://secure.netflix.com/en_us/layout/ecweb/netflix-app-icon_152.jpg", Width: 152, Height: 152, Format: "jpeg", Bytes: 5184, Sha1sum: "0edf41de3b2abf4c3135a564a6d47b87fab517e0"},
		{URL: "https://secure.netflix.com/en_us/icons/nficon2014.4.ico", Width: 48, Height: 48, Format: "ico", Bytes: 26306, Sha1sum: "b76e2bc10f53e3ac9ee677ea5d503e10355da6db"},
		{URL: "https://www.netflix.com/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 1150, Sha1sum: "ffeb019da4d7fabc0dd55184396e3d3b4a335e23"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestGoogleWithOnePixelGifs(t *testing.T) {
	actualImages, err := fetchIconsWithVCR("google.vcr", "google.com")
	check(err)
	expectedImages := []Icon{
		{URL: "http://www.google.de/favicon.ico?gfe_rd=cr&ei=JSbJVO7AFaeh8wfo-oHIAQ", Width: 32, Height: 32, Format: "ico", Bytes: 5430, Error: error(nil), Sha1sum: "e2020bf4f2b65f62434c62ea967973140b3300df"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestGithubWithIconHrefLinks(t *testing.T) {
	actualImages, err := fetchIconsWithVCR("github.vcr", "github.com")
	check(err)
	expectedImages := []Icon{
		{URL: "https://github.com/apple-touch-icon-144.png", Width: 144, Height: 144, Format: "png", Bytes: 796, Sha1sum: "2626d8f64d5d3a76bd535151dfe84b62d3f3ee63"},
		{URL: "https://github.com/apple-touch-icon.png", Width: 120, Height: 120, Format: "png", Bytes: 676, Sha1sum: "8eb0b1d3f0797c0fe94368f4ad9a2c9513541cd2"},
		{URL: "https://github.com/apple-touch-icon-114.png", Width: 114, Height: 114, Format: "png", Bytes: 648, Sha1sum: "644982478322a731a6bd8fe7fad9afad8f4a3c4b"},
		{URL: "https://github.com/apple-touch-icon-precomposed.png", Width: 57, Height: 57, Format: "png", Bytes: 705, Sha1sum: "f97e9954be83f93ce2a1a85c2d8f93e2235c887f"},
		{URL: "https://assets-cdn.github.com/favicon.ico", Width: 32, Height: 32, Format: "ico", Bytes: 6518, Sha1sum: "4eda7c0f3a36181f483dd0a14efe9f58c8b29814"},
		{URL: "https://github.com/favicon.ico", Width: 32, Height: 32, Format: "ico", Bytes: 6518, Sha1sum: "4eda7c0f3a36181f483dd0a14efe9f58c8b29814"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestARDWithSortBySize(t *testing.T) {
	actualImages, err := fetchIconsWithVCR("ard.vcr", "ard.de")
	check(err)
	expectedImages := []Icon{
		{URL: "http://www.ard.de/ARD-144.png", Width: 144, Height: 144, Format: "png", Bytes: 29228, Sha1sum: "a6be15435a80e9de7978d203a3f2723940ab6bda"},
		{URL: "http://www.ard.de/apple-touch-icon-precomposed.png", Width: 144, Height: 144, Format: "png", Bytes: 29228, Sha1sum: "a6be15435a80e9de7978d203a3f2723940ab6bda"},
		{URL: "http://www.ard.de/apple-touch-icon.png", Width: 144, Height: 144, Format: "png", Bytes: 29228, Sha1sum: "a6be15435a80e9de7978d203a3f2723940ab6bda"},
		{URL: "http://www.ard.de/favicon.ico", Width: 144, Height: 144, Format: "ico", Bytes: 116094, Sha1sum: "e5bd22dda5647420c5d2053ee9fd21b543dc09a8"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestParsingNotHTML(t *testing.T) {
	actualImages, err := fetchIconsWithVCR("not_image.vcr", "http://wikipedia.org/favicon.ico")

	assertEquals(t, 0, len(actualImages))
	assertEquals(t, errors.New("besticon: could not parse html"), err)
}

func TestParsingEmptyResponse(t *testing.T) {
	actualImages, err := fetchIconsWithVCR("empty_body.vcr", "foobar.com")

	assertEquals(t, 0, len(actualImages))
	assertEquals(t, errors.New("besticon: empty response"), err)
}

func mustFindIconLinks(html []byte) []string {
	links, e := findIcons(html)
	check(e)
	return links
}

func TestFindBestIcon(t *testing.T) {
	i, err := FetchBestIcon("github.com")
	assertEquals(t, nil, err)
	assertEquals(t, &Icon{URL: "https://github.com/apple-touch-icon-144.png", Width: 144, Height: 144, Format: "png", Bytes: 796, Error: error(nil), Sha1sum: "2626d8f64d5d3a76bd535151dfe84b62d3f3ee63"}, i)
}

func TestFindBestIconNoIcons(t *testing.T) {
	_, err := FetchBestIcon("example.com")
	assertEquals(t, errors.New("besticon: no icons found for site"), err)
}

func TestLinkExtraction(t *testing.T) {
	assertEquals(t, []string{"/graphics/favicon.ico?v=005",
		"/graphics/apple-touch-icon.png"},
		mustFindIconLinks(mustReadFile("testdata/daringfireball.html")))

	assertEquals(t, []string{"/wp-content/assets/dist/img/icon/favicon.ico",
		"/wp-content/assets/dist/img/icon/apple-touch-icon.png",
		"/wp-content/assets/dist/img/icon/apple-touch-icon-precomposed.png",
		"/wp-content/assets/dist/img/icon/apple-touch-icon-57x57-precomposed.png",
		"/wp-content/assets/dist/img/icon/apple-touch-icon-114x114-precomposed.png",
		"/wp-content/assets/dist/img/icon/apple-touch-icon-144x144-precomposed.png"},
		mustFindIconLinks(mustReadFile("testdata/newyorker.html")))
}

func TestImageSizeDetection(t *testing.T) {
	assertEquals(t, 1, getImageWidthForFile("testdata/pixel.png"))
	assertEquals(t, 1, getImageWidthForFile("testdata/pixel.gif"))
	assertEquals(t, 199, getImageWidthForFile("testdata/mat.jpg"))
	assertEquals(t, 48, getImageWidthForFile("testdata/favicon.ico"))
	assertEquals(t, 16, getImageWidthForFile("testdata/rose.bmp"))
}

const testdataDir = "testdata/"

func fetchIconsWithVCR(vcrFile string, url string) ([]Icon, error) {
	path := testdataDir + vcrFile
	f, err := os.Open(path)

	if err != nil {
		return recordFetchIcons(path, url)
	}

	defer f.Close()
	return replayFetchIcons(path, url)
}

func recordFetchIcons(vcrFile string, url string) ([]Icon, error) {
	file, err := os.Create(vcrFile)
	check(err)
	defer file.Close()

	gzfile := gzip.NewWriter(file)
	defer gzfile.Close()

	client := vcr.NewRecordingClient(gzfile)
	icons, err := fetchIconsWithClient(url, &client)
	return icons, err
}

func replayFetchIcons(vcrFile string, url string) ([]Icon, error) {
	gzfile, err := os.Open(vcrFile)
	check(err)
	defer gzfile.Close()

	file, err := gzip.NewReader(gzfile)
	check(err)
	defer file.Close()

	client, err := vcr.NewReplayerClient(file)
	check(err)
	icons, err := fetchIconsWithClient(url, &client)
	return icons, err
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
