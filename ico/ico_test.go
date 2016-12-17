package ico

import (
	"errors"
	"fmt"
	"image"
	"os"
	"reflect"
	"testing"
)

func TestParseICO(t *testing.T) {
	assertEquals(t, 3, GetNumberOfIconsInFile(t, "favicon.ico"))
	assertDecodesImage(t, "github.ico")
	assertDecodesImage(t, "besticon.ico")
	assertDecodesImage(t, "addthis.ico")
	assertDecodesImage(t, "wowhead.ico")
}

func TestParseICODetails(t *testing.T) {
	entries := []icondirEntry{
		{Width: 0x30, Height: 0x30, PaletteCount: 0x10, Reserved: 0x0, ColorPlanes: 0x1, BitsPerPixel: 0x4, Size: 0x668, Offset: 0x36},
		{Width: 0x20, Height: 0x20, PaletteCount: 0x10, Reserved: 0x0, ColorPlanes: 0x1, BitsPerPixel: 0x4, Size: 0x2e8, Offset: 0x69e},
		{Width: 0x10, Height: 0x10, PaletteCount: 0x10, Reserved: 0x0, ColorPlanes: 0x1, BitsPerPixel: 0x4, Size: 0x128, Offset: 0x986},
	}
	assertEquals(t, icondir{Reserved: 0, Type: 1, Count: 3, Entries: entries}, mustParseIcoFile(t, "favicon.ico"))
}

func TestFindBestIcon(t *testing.T) {
	dir := mustParseIcoFile(t, "favicon.ico")
	best := dir.FindBestIcon()

	assertEquals(t,
		&icondirEntry{Width: 0x30, Height: 0x30, PaletteCount: 0x10, Reserved: 0x0, ColorPlanes: 0x1, BitsPerPixel: 0x4, Size: 0x668, Offset: 0x36},
		best)
}

func TestColorCount(t *testing.T) {
	dir := mustParseIcoFile(t, "favicon.ico")
	best := dir.FindBestIcon()
	assertEquals(t, 16, best.ColorCount())

	dir = mustParseIcoFile(t, "github.ico")
	best = dir.FindBestIcon()
	assertEquals(t, 256, best.ColorCount())
}

func TestDecodeConfig(t *testing.T) {
	f, err := os.Open("favicon.ico")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	imageConfig, _, err := image.DecodeConfig(f)

	assertEquals(t, nil, err)
	assertEquals(t,
		image.Config{ColorModel: nil, Width: 48, Height: 48},
		imageConfig)
}

func TestDecodeConfigWithBrokenIco(t *testing.T) {
	f, err := os.Open("broken.ico")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	imageConfig, _, err := image.DecodeConfig(f)

	assertEquals(t,
		image.Config{},
		imageConfig)

	assertEquals(t, errors.New("unexpected EOF"), err)
}

func TestParse256WidthHeightIco(t *testing.T) {
	assertEquals(t, 5, GetNumberOfIconsInFile(t, "codeplex.ico"))

	f, err := os.Open("codeplex.ico")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	imageConfig, _, err := image.DecodeConfig(f)

	assertEquals(t, nil, err)
	assertEquals(t,
		image.Config{ColorModel: nil, Width: 256, Height: 256},
		imageConfig)

	assertDecodesImage(t, "codeplex.ico")
}

func GetNumberOfIconsInFile(t *testing.T, filename string) int {
	dir := mustParseIcoFile(t, filename)
	return int(dir.Count)
}

func parseIcoFile(filename string) (*icondir, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseIco(f)
}

func mustParseIcoFile(t *testing.T, filename string) icondir {
	dir, err := parseIcoFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	return *dir
}

func assertEquals(t *testing.T, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		fail(t, fmt.Sprintf("Not equal: %#v (expected)\n"+
			"        != %#v (actual)", expected, actual))
	}
}

func assertDecodesImage(t *testing.T, file string) {
	f, err := os.Open(file)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = Decode(f)
	assertEquals(t, nil, err)
}

func fail(t *testing.T, failureMessage string) {
	t.Errorf("\t%s\n"+
		"\r\t",
		failureMessage)
}
