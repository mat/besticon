package colorfinder

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestSimplePixel(t *testing.T) {
	assertFindsRightColor(t, "white1x1.png", "ffffff")
	assertFindsRightColor(t, "black1x1.png", "000000")
}

func TestImageFormats(t *testing.T) {
	assertFindsRightColor(t, "white1x1.png", "ffffff")
	assertFindsRightColor(t, "white1x1.gif", "ffffff")
}

func TestFindColors01(t *testing.T) {
	assertFindsRightColor(t, "icon01.png.gz", "113671")
}

func TestFindColors02(t *testing.T) {
	assertFindsRightColor(t, "icon02.png.gz", "cb1c1f")
}

func TestFindColors03(t *testing.T) {
	assertFindsRightColor(t, "icon03.png.gz", "f48024")
}

func TestFindColors04(t *testing.T) {
	assertFindsRightColor(t, "icon04.png.gz", "cfdc00")
}

func TestFindColors05(t *testing.T) {
	assertFindsRightColor(t, "icon05.png.gz", "ffa700")
}

func TestFindColors06(t *testing.T) {
	assertFindsRightColor(t, "icon06.png.gz", "ff6600")
}

func TestFindColors07(t *testing.T) {
	assertFindsRightColor(t, "icon07.png.gz", "e61a30")
}

func TestFindColors08(t *testing.T) {
	// .ico with png
	assertFindsRightColor(t, "icon08.ico.gz", "14e06e")
}

func TestFindColors09(t *testing.T) {
	// .ico with 32-bit bmp
	assertFindsRightColor(t, "icon09.ico.gz", "1c5182")
}

func TestFindColors10(t *testing.T) {
	// .ico with 8-bit bmp, ColorsUsed=0
	assertFindsRightColor(t, "icon10.ico.gz", "fe6d4c")
}

func TestFindColors11(t *testing.T) {
	// .ico with 8-bit bmp, ColorsUsed=256
	assertFindsRightColor(t, "icon11.ico.gz", "a30000")
}

func BenchmarkFindMainColor152x152(b *testing.B) {
	file, _ := os.Open(testdataDir + "icon02.png.gz")
	gzReader, _ := gzip.NewReader(file)
	byts, _ := ioutil.ReadAll(gzReader)
	imgReader := bytes.NewReader(byts)
	img, _, err := image.Decode(imgReader)
	if err != nil {
		log.Fatal(err)
	}

	b.ResetTimer()

	cf := ColorFinder{}
	for i := 0; i < b.N; i++ {
		col, err := cf.FindMainColor(img)
		if err != nil {
			b.Errorf("Unexpected error:  %#v", err)
		}
		if ColorToHex(col) != "cb1c1f" {
			b.Errorf("Wrong color: %s", ColorToHex(col))
		}

		imgReader.Seek(0, 0)
	}
}

func BenchmarkFindMainColor57x57(b *testing.B) {
	file, _ := os.Open(testdataDir + "icon07.png.gz")
	gzReader, _ := gzip.NewReader(file)
	byts, _ := ioutil.ReadAll(gzReader)
	imgReader := bytes.NewReader(byts)
	img, _, err := image.Decode(imgReader)
	if err != nil {
		log.Fatal(err)
	}

	b.ResetTimer()

	cf := ColorFinder{}
	for i := 0; i < b.N; i++ {
		col, err := cf.FindMainColor(img)
		if err != nil {
			b.Errorf("Unexpected error:  %#v", err)
		}
		if ColorToHex(col) != "e61a30" {
			b.Errorf("Wrong color: %s", ColorToHex(col))
		}

		imgReader.Seek(0, 0)
	}
}

const testdataDir = "testdata/"

func assertFindsRightColor(t *testing.T, fileName string, expectedHexColor string) {
	var imgReader io.ReadCloser

	path := testdataDir + fileName
	imgReader, err := os.Open(path)
	check(t, err)

	if strings.HasSuffix(path, ".gz") {
		imgReader, err = gzip.NewReader(imgReader)
		check(t, err)
	}

	defer imgReader.Close()
	img, _, err := image.Decode(imgReader)
	if err != nil {
		log.Fatal(err)
	}

	cf := ColorFinder{}
	actualColor, err := cf.FindMainColor(img)
	check(t, err)

	assertEquals(t, expectedHexColor, ColorToHex(actualColor))
}

func check(t *testing.T, err error) {
	if err != nil {
		fail(t, fmt.Sprintf("Unexpected error:  %#v", err))
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
