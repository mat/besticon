package lettericon

import (
	"bytes"
	"fmt"
	"image/color"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestColorFromHex(t *testing.T) {
	assertColor(t, "#000000", &color.RGBA{0, 0, 0, 0xff})
	assertColor(t, "#ffffff", &color.RGBA{255, 255, 255, 0xff})
	assertColor(t, "#dfdfdf", &color.RGBA{223, 223, 223, 0xff})
}

func TestRender(t *testing.T) {
	assertCorrectImageData(t, "A", 16, "123456")
	assertCorrectImageData(t, "X", 32, "dfdfdf")
}

func TestPickForegroundColor(t *testing.T) {
	orange := color.RGBA{255, 152, 0, 255}
	assertEquals(t, color.White, pickForegroundColor(orange))

	white := color.RGBA{254, 255, 252, 255}
	assertEquals(t, lightDark, pickForegroundColor(white))
}

func TestContrastRatio(t *testing.T) {
	white := &color.RGBA{255, 255, 255, 0}
	blue := &color.RGBA{0, 0, 255, 0}

	assertFloatEquals(t, 1.0, contrastRatio(white, white))
	assertFloatEquals(t, 1.0, contrastRatio(blue, blue))
	assertFloatEquals(t, 8.6, contrastRatio(white, blue))
}

func TestRelativeLuminance(t *testing.T) {
	white := &color.RGBA{255, 255, 255, 0}
	blue := &color.RGBA{0, 0, 255, 0}

	assertFloatEquals(t, 1.0, relativeLuminance(white))
	assertFloatEquals(t, 0.07, relativeLuminance(blue))
}

const testdataDir = "testdata/"

func assertCorrectImageData(t *testing.T, letter string, width int, hexColor string) {
	imageData, err := renderBytes(letter, mustColorFromHex(hexColor), width)
	if err != nil {
		fail(t, fmt.Sprintf("could not generate icon: %s", err))
		return
	}

	// "A-144-123456.png"
	file := fmt.Sprintf(testdataDir+"%s-%d-%s.png", letter, width, hexColor)
	fileData, err := bytesFromFile(file)
	if err != nil {
		fail(t, fmt.Sprintf("could not load icon file: %s", err))
		return
	}

	assertEquals(t, len(fileData), len(imageData))
	assertEquals(t, fileData, imageData)
}

func mustColorFromHex(hexColor string) color.Color {
	col, err := ColorFromHex(hexColor)
	if err != nil {
		panic(err)
	}
	return col
}

func BenchmarkColorFromHex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ColorFromHex("#dfdfdf")
	}
}

func bytesFromFile(file string) ([]byte, error) {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return dat, nil
}

func renderBytes(letter string, bgColor color.Color, width int) ([]byte, error) {
	var b bytes.Buffer

	err := Render(letter, bgColor, width, &b)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func BenchmarkRender(b *testing.B) {
	Render("X", DefaultBackgroundColor, 144, ioutil.Discard) // warmup
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		Render("X", DefaultBackgroundColor, 144, ioutil.Discard)
	}
}

func TestMainLetterFromURL(t *testing.T) {
	assertEquals(t, "b", MainLetterFromURL("http://better-idea.org"))
	assertEquals(t, "b", MainLetterFromURL("http://www.better-idea.org"))
	assertEquals(t, "b", MainLetterFromURL("better-idea.org"))
	assertEquals(t, "b", MainLetterFromURL("www.better-idea.org"))

	assertEquals(t, "s", MainLetterFromURL("some-user.blogspot.com"))
	assertEquals(t, "b", MainLetterFromURL("blogspot.com"))
}

func TestIconPath(t *testing.T) {
	assertEquals(t, "/lettericons/A-120-000000.png", IconPath("a", "120", &color.RGBA{0, 0, 0, 0}))
	assertEquals(t, "/lettericons/Z-100-640ac8.png", IconPath("z", "100", &color.RGBA{100, 10, 200, 0}))
}

func TestParseIconPath(t *testing.T) {
	var char string
	var col *color.RGBA
	var size int

	char, _, _ = ParseIconPath("lettericons/")
	assertEquals(t, "", char)

	char, _, _ = ParseIconPath("lettericons/A")
	assertEquals(t, "A", char)

	char, _, _ = ParseIconPath("lettericons/B.png")
	assertEquals(t, "B", char)

	char, _, size = ParseIconPath("lettericons/C-120.png")
	assertEquals(t, "C", char)
	assertEquals(t, 120, size)

	char, col, size = ParseIconPath("lettericons/D-150-ababab.png")
	assertEquals(t, "D", char)
	assertEquals(t, 150, size)
	assertEquals(t, &color.RGBA{171, 171, 171, 0xff}, col)
}

func assertColor(t *testing.T, hexColor string, expectedColor color.Color) {
	actualColor, err := ColorFromHex(hexColor)
	if err != nil {
		fail(t, fmt.Sprintf("'%s' does not into a color", hexColor))
		return
	}

	assertEquals(t, expectedColor, actualColor)
}
func assertEquals(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		fail(t, fmt.Sprintf("Not equal: %v (expected)\n"+
			"        != %v (actual)", expected, actual))
	}
}

func assertFloatEquals(t *testing.T, expected, actual float64) {
	delta := expected - actual
	if delta > 0.01 {
		fail(t, fmt.Sprintf("Not equal: %v (expected)\n"+
			"        != %v (actual)", expected, actual))
	}
}

func fail(t *testing.T, failureMessage string) {
	t.Errorf("\t%s\n"+
		"\r\t",
		failureMessage)
}
