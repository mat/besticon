package besticon

import (
	"errors"
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

func TestKicktipp(t *testing.T) {
	actualImages, finder, err := fetchIconsWithVCR("kicktipp.vcr", "http://kicktipp.de")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		{URL: "http://info.kicktipp.de/assets/img/jar_cb333387130/assets/img/logos/apple-touch-icon-57x57-precomposed.png", Width: 57, Height: 57, Format: "png", Bytes: 1535, Sha1sum: "79aae9e0df7d52ed50ac47c1dd4bd16e2ddf8b4a"},
		{URL: "http://www.kicktipp.de/apple-touch-icon-precomposed.png", Width: 57, Height: 57, Format: "png", Bytes: 1622, Sha1sum: "fd4306aefd9ed2b4983697ca56301d8810a03987"},
		{URL: "http://www.kicktipp.de/apple-touch-icon.png", Width: 57, Height: 57, Format: "png", Bytes: 1622, Sha1sum: "fd4306aefd9ed2b4983697ca56301d8810a03987"},
		{URL: "http://www.kicktipp.de/favicon.ico", Width: 32, Height: 32, Format: "gif", Bytes: 35275, Sha1sum: "09297d0ffe17149c3d4d4a3a3a8c7e8c51932d58"},
		{URL: "http://info.kicktipp.de/assets/img/jar_cb1652512069/assets/img/logos/favicon.png", Width: 16, Height: 16, Format: "png", Bytes: 1820, Sha1sum: "04b49fac810828f6723cd763600af23f0edbde03"},
	}
	assertEquals(t, expectedImages, actualImages)

	actualImage := finder.IconInSizeRange(SizeRange{20, 80, 500})
	expectedImage := &Icon{URL: "http://info.kicktipp.de/assets/img/jar_cb333387130/assets/img/logos/apple-touch-icon-57x57-precomposed.png", Width: 57, Height: 57, Format: "png", Bytes: 1535, Sha1sum: "79aae9e0df7d52ed50ac47c1dd4bd16e2ddf8b4a"}
	assertEquals(t, expectedImage, actualImage)
}

func TestDaringfireball(t *testing.T) {
	actualImages, finder, err := fetchIconsWithVCR("daringfireball.net.vcr", "http://daringfireball.net")
	assertEquals(t, nil, err)

	expectedImages := []Icon{
		{URL: "http://daringfireball.net/graphics/apple-touch-icon.png", Width: 314, Height: 314, Format: "png", Bytes: 8642, Error: error(nil), Sha1sum: "f47cf7cf13ec1a74049d99d9f1565354b5b20317"},
		{URL: "http://daringfireball.net/favicon.ico", Width: 32, Height: 32, Format: "ico", Bytes: 6518, Error: error(nil), Sha1sum: "c066c70aa1dd2b4347d3095592aac28b55e85d04"},
		{URL: "http://daringfireball.net/graphics/favicon.ico?v=005", Width: 32, Height: 32, Format: "ico", Bytes: 6518, Error: error(nil), Sha1sum: "c066c70aa1dd2b4347d3095592aac28b55e85d04"},
	}

	assertEquals(t, expectedImages, actualImages)

	actualImage := finder.IconInSizeRange(SizeRange{20, 80, 500})
	expectedImage := &Icon{URL: "http://daringfireball.net/graphics/apple-touch-icon.png", Width: 314, Height: 314, Format: "png", Bytes: 8642, Error: error(nil), Sha1sum: "f47cf7cf13ec1a74049d99d9f1565354b5b20317"}
	assertEquals(t, expectedImage, actualImage)
}

func TestAwsAmazonChangingBaseURL(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("aws.amazon.vcr", "http://aws.amazon.com")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		{URL: "http://a0.awsstatic.com/main/images/site/touch-icon-ipad-144-precomposed.png", Width: 144, Height: 144, Format: "png", Bytes: 3944, Sha1sum: "225817df40ff11d083c282d08b49a5ed50fd0f2d"},
		{URL: "http://a0.awsstatic.com/main/images/site/touch-icon-iphone-114-precomposed.png", Width: 114, Height: 114, Format: "png", Bytes: 3081, Sha1sum: "58aabb2a99fcb283710fd200c9feed69c015a29e"},
		{URL: "http://a0.awsstatic.com/main/images/site/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 1150, Sha1sum: "a64f3616e3671b835f8b208b7339518d8b386a08"},
		{URL: "http://aws.amazon.com/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 1150, Sha1sum: "a64f3616e3671b835f8b208b7339518d8b386a08"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestNetflixWitCookieRedirects(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("netflix.vcr", "http://netflix.com")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		{URL: "https://assets.nflxext.com/us/ffe/siteui/common/icons/nficon2016.png", Width: 64, Height: 64, Format: "png", Bytes: 1755, Sha1sum: "867e51e9b4a474c19da52d6454076c007a9d01f2"},
		{URL: "https://assets.nflxext.com/us/ffe/siteui/common/icons/nficon2016.ico", Width: 64, Height: 64, Format: "ico", Bytes: 16958, Sha1sum: "931e18dfc6e7d950dc2f2bbdfe31e1ea720acf7c"},
		{URL: "https://www.netflix.com/favicon.ico", Width: 64, Height: 64, Format: "ico", Bytes: 16958, Sha1sum: "931e18dfc6e7d950dc2f2bbdfe31e1ea720acf7c"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestAolWithOnePixelGifs(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("aol.vcr", "http://aol.com")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		{URL: "http://www.aol.com/favicon.ico", Width: 32, Height: 32, Format: "ico", Bytes: 7886, Error: error(nil), Sha1sum: "c474f8307362367be553b884878e221f25fcb80b"},
		{URL: "http://www.aol.com/favicon.ico?v=2", Width: 32, Height: 32, Format: "ico", Bytes: 7886, Error: error(nil), Sha1sum: "c474f8307362367be553b884878e221f25fcb80b"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestGithubWithIconHrefLinks(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("github.vcr", "http://github.com")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		// later - for svg
		// {URL: "https://assets-cdn.github.com/pinned-octocat.svg", Width: 9999, Height: 9999, Format: "svg", Bytes: 1009, Sha1sum: "5dec89dc9dbcebff4a2aeddb9426c7a8e237ca16"},
		{URL: "https://github.com/apple-touch-icon-144.png", Width: 144, Height: 144, Format: "png", Bytes: 796, Sha1sum: "2626d8f64d5d3a76bd535151dfe84b62d3f3ee63"},
		{URL: "https://github.com/apple-touch-icon.png", Width: 120, Height: 120, Format: "png", Bytes: 676, Sha1sum: "8eb0b1d3f0797c0fe94368f4ad9a2c9513541cd2"},
		{URL: "https://github.com/apple-touch-icon-114.png", Width: 114, Height: 114, Format: "png", Bytes: 648, Sha1sum: "644982478322a731a6bd8fe7fad9afad8f4a3c4b"},
		{URL: "https://github.com/apple-touch-icon-precomposed.png", Width: 57, Height: 57, Format: "png", Bytes: 705, Sha1sum: "f97e9954be83f93ce2a1a85c2d8f93e2235c887f"},
		{URL: "https://assets-cdn.github.com/favicon.ico", Width: 32, Height: 32, Format: "ico", Bytes: 6518, Sha1sum: "4eda7c0f3a36181f483dd0a14efe9f58c8b29814"},
		{URL: "https://github.com/favicon.ico", Width: 32, Height: 32, Format: "ico", Bytes: 6518, Sha1sum: "4eda7c0f3a36181f483dd0a14efe9f58c8b29814"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestEat24WithBaseTag(t *testing.T) {
	actualImages, finder, err := fetchIconsWithVCR("eat24.vcr", "http://eat24.com")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		// later - for svg
		// {URL: "http://eat24hours.com/static/v4/images/favicon.svg", Width: 9999, Height: 9999, Format: "svg", Bytes: 1498, Sha1sum: "db580998de6dd01e4433865b5f77bd6491bbc7bc"},
		{URL: "http://eat24hours.com/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 1406, Sha1sum: "f8914a1135e718b11cc93b7a362655ca358c16fb"},
	}
	assertEquals(t, expectedImages, actualImages)

	actualImage := finder.IconInSizeRange(SizeRange{120, 150, 500})
	assertEquals(t, (*Icon)(nil), actualImage)
}

func TestCar2goWithRelativeURL(t *testing.T) {
	// ../../assets/icon.ico
	actualImages, finder, err := fetchIconsWithVCR("car2go.vcr", "http://car2go.com")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		{URL: "https://www.car2go.com/media/assets/patterns/static/img/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 1150, Sha1sum: "860e9ef188675f4f0b7036c2d22e6497ea732282"},
	}
	assertEquals(t, expectedImages, actualImages)

	actualImage := finder.IconInSizeRange(SizeRange{80, 120, 200})
	assertEquals(t, (*Icon)(nil), actualImage)
}

func TestAlibabaWithBaseTagWithoutScheme(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("alibaba.vcr", "http://alibaba.com")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		{URL: "http://is.alicdn.com/simg/single/icon/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 1406, Sha1sum: "4ffbef9b6044c62cd6c8b1ee0913ba93e6e80072"},
		{URL: "http://www.alibaba.com/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 1406, Sha1sum: "4ffbef9b6044c62cd6c8b1ee0913ba93e6e80072"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestDnevnikWithCapitalizedIconTag(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("dnevnik.vcr", "http://www.dnevnik.bg")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		{URL: "http://www.dnevnik.bg/images/layout/apple-touch-icon.png", Width: 180, Height: 180, Format: "png", Bytes: 1597, Sha1sum: "16af14e168879ac52f594c67b4298f76d768a5eb"},
		{URL: "http://www.dnevnik.bg/apple-touch-icon.png", Width: 129, Height: 129, Format: "png", Bytes: 2092, Sha1sum: "f96615ddf0d9e75e28b7420ed10bbdc1de6f6dab"},
		{URL: "http://www.dnevnik.bg/favicon.ico", Width: 32, Height: 32, Format: "ico", Bytes: 6518, Sha1sum: "72b4cb7ca529a5d3f5ebf380e77108bd2c04bc04"},
		{URL: "http://www.dnevnik.bg/images/layout/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 894, Sha1sum: "acf6cacab957c263851e8c13ea68ad8ecb5fcb94"},
	}
	assertEquals(t, expectedImages, actualImages)
}

func TestARDWithSortBySize(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("ard.vcr", "http://ard.de")
	assertEquals(t, nil, err)
	expectedImages := []Icon{
		{URL: "http://www.ard.de/ARD-144.png", Width: 144, Height: 144, Format: "png", Bytes: 29228, Sha1sum: "a6be15435a80e9de7978d203a3f2723940ab6bda"},
		{URL: "http://www.ard.de/apple-touch-icon-precomposed.png", Width: 144, Height: 144, Format: "png", Bytes: 29228, Sha1sum: "a6be15435a80e9de7978d203a3f2723940ab6bda"},
		{URL: "http://www.ard.de/apple-touch-icon.png", Width: 144, Height: 144, Format: "png", Bytes: 29228, Sha1sum: "a6be15435a80e9de7978d203a3f2723940ab6bda"},
		{URL: "http://www.ard.de/favicon.ico", Width: 144, Height: 144, Format: "ico", Bytes: 116094, Sha1sum: "e5bd22dda5647420c5d2053ee9fd21b543dc09a8"},
	}

	assertEquals(t, expectedImages, actualImages)
}

func TestMortenmøllerWithIDNAHost(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("mortenmoller.vcr", "https://mortenmøller.dk")
	assertEquals(t, nil, err)
	assertEquals(t, 13, len(actualImages))
}

func TestYoutubeWithDomainRewrite(t *testing.T) {
	// This test can only work because with HostOnlyDomains accordingly
	_, finder, err := fetchIconsWithVCR("youtube.vcr", "http://youtube.com/does-not-exist")
	ico := finder.IconInSizeRange(SizeRange{0, 80, 200})
	assertEquals(t, &Icon{URL: "https://s.ytimg.com/yts/img/favicon_96-vfldSA3ca.png", Width: 96, Height: 96, Format: "png", Bytes: 1510, Sha1sum: "7149bef987538d34e2ab6e069d08211d0a6e407d"}, ico)
	assertEquals(t, nil, err)
}

func TestRandomOrg(t *testing.T) {
	// https://github.com/mat/besticon/issues/28
	_, finder, err := fetchIconsWithVCR("random.org.vcr", "https://random.org")
	assertEquals(t, nil, err)

	actualImage := finder.IconInSizeRange(SizeRange{16, 32, 64})
	expectedImage := &Icon{URL: "https://www.random.org/favicon.ico", Width: 16, Height: 16, Format: "ico", Bytes: 2550, Error: error(nil), Sha1sum: "f8087e651b79c36d206f6f408d7fe74dcb11d351"}
	assertEquals(t, expectedImage, actualImage)
}

func TestArchiveOrgWithJpg(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("archive.org.vcr", "https://archive.org")
	assertEquals(t, nil, err)

	expectedImages := []Icon{
		{URL: "https://archive.org/apple-touch-icon-precomposed.png", Width: 180, Height: 180, Format: "png", Bytes: 5495, Sha1sum: "7b583a9eee7c4f705f2a93ddc50e6927bde4b634"},
		{URL: "https://archive.org/apple-touch-icon.png", Width: 180, Height: 180, Format: "png", Bytes: 7494, Sha1sum: "093d651bf04e480e5c25167e780455c879c02447"},
		{URL: "https://archive.org/images/glogo.jpg", Width: 40, Height: 40, Format: "jpg", Bytes: 3213, Sha1sum: "279fe766b791ae83a10765a8790a0928448a4e35"},
		{URL: "https://archive.org/favicon.ico", Width: 32, Height: 32, Format: "ico", Bytes: 4286, Sha1sum: "b18786d77997511ab0f6e5c9d3c5b9e1bff164be"},
	}
	assertEquals(t, expectedImages, actualImages)
}

func TestParsingInexistentSite(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("not_existent.vcr", "http://wikipedia.org/does-not-exist")

	assertEquals(t, errors.New("besticon: not found"), err)
	assertEquals(t, 0, len(actualImages))
}

func TestParsingEmptyResponse(t *testing.T) {
	actualImages, _, err := fetchIconsWithVCR("empty_body.vcr", "http://foobar.com")

	assertEquals(t, 0, len(actualImages))
	assertEquals(t, errors.New("besticon: empty response"), err)
}

func TestMainColorForIconsWithBrokenImageData(t *testing.T) {
	icn := Icon{Format: "png", ImageData: []byte("broken-image-data")}
	colr := MainColorForIcons([]Icon{icn})
	assertEquals(t, (*color.RGBA)(nil), colr)
}

func TestFindBestIconNoIcons(t *testing.T) {
	icons, _, _ := fetchIconsWithVCR("example.com.vcr", "http://example.com")
	assertEquals(t, 0, len(icons))
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

const testdataDir = "testdata/"

func fetchIconsWithVCR(vcrFile string, url string) ([]Icon, *IconFinder, error) {
	path := testdataDir + vcrFile
	c, f, err := vcr.Client(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	setHTTPClient(c)
	finder := IconFinder{}
	finder.HostOnlyDomains = []string{"youtube.com"}
	icons, e := finder.FetchIcons(url)
	return icons, &finder, e
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
