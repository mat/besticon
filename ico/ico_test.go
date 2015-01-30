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
	assertEquals(t, 3, GetNumberOfIconsInFile("favicon.ico"))
}

func TestParseICODetails(t *testing.T) {
	entries := []icondirEntry{
		icondirEntry{Width: 0x30, Height: 0x30, Colors: 0x10, Reserved: 0x0, Planes: 0x1, Bits: 0x4, Bytes: 0x668, Offset: 0x36},
		icondirEntry{Width: 0x20, Height: 0x20, Colors: 0x10, Reserved: 0x0, Planes: 0x1, Bits: 0x4, Bytes: 0x2e8, Offset: 0x69e},
		icondirEntry{Width: 0x10, Height: 0x10, Colors: 0x10, Reserved: 0x0, Planes: 0x1, Bits: 0x4, Bytes: 0x128, Offset: 0x986},
	}
	assertEquals(t, icondir{Reserved: 0, Type: 1, Count: 3, Entries: entries}, mustParseIcoFile("favicon.ico"))
}

func TestFindBestIcon(t *testing.T) {
	dir := mustParseIcoFile("favicon.ico")
	best := dir.FindBestIcon()

	assertEquals(t,
		icondirEntry{Width: 0x30, Height: 0x30, Colors: 0x10, Reserved: 0x0, Planes: 0x1, Bits: 0x4, Bytes: 0x668, Offset: 0x36},
		best)
}

func TestColorCount(t *testing.T) {
	dir := mustParseIcoFile("favicon.ico")
	best := dir.FindBestIcon()
	assertEquals(t, 16, best.ColorCount())

	dir = mustParseIcoFile("github.ico")
	best = dir.FindBestIcon()
	assertEquals(t, 256, best.ColorCount())
}

func TestDecodeConfig(t *testing.T) {
	f, err := os.Open("favicon.ico")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	imageConfig, _, err := image.DecodeConfig(f)

	assertEquals(t,
		image.Config{ColorModel: nil, Width: 48, Height: 48},
		imageConfig)
}

func TestDecodeConfigWithBrokenIco(t *testing.T) {
	f, err := os.Open("broken.ico")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	imageConfig, _, err := image.DecodeConfig(f)

	assertEquals(t,
		image.Config{},
		imageConfig)

	assertEquals(t, errors.New("unexpected EOF"), err)
}

func GetFirstIconInFile(filename string) icondirEntry {
	dir := mustParseIcoFile(filename)
	return dir.Entries[0]
}

func GetNumberOfIconsInFile(filename string) int {
	dir := mustParseIcoFile(filename)
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

func mustParseIcoFile(filename string) icondir {
	dir, err := parseIcoFile(filename)
	if err != nil {
		panic(err)
	}

	return *dir
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
