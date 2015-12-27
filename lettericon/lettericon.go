package lettericon

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"math"
	"net/url"
	"path"
	"strconv"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"golang.org/x/net/publicsuffix"

	"github.com/golang/freetype/truetype"
	"github.com/mat/besticon/colorfinder"
	"github.com/mat/besticon/lettericon/fonts"
)

const dpi = 72

const fontSizeFactor = 0.6180340     // (by taste)
const yOffsetFactor = 102.0 / 1024.0 // (by trial and error) :-)

func Render(letter string, bgColor color.Color, width int, out io.Writer) error {
	fg := pickForegroundColor(bgColor)

	rgba := image.NewRGBA(image.Rect(0, 0, width, width))
	draw.Draw(rgba, rgba.Bounds(), &image.Uniform{bgColor}, image.ZP, draw.Src)

	fontSize := fontSizeFactor * float64(width)
	d := &font.Drawer{
		Dst: rgba,
		Src: &image.Uniform{fg},
		Face: truetype.NewFace(fnt, &truetype.Options{
			Size:    fontSize,
			DPI:     dpi,
			Hinting: font.HintingNone,
		}),
	}

	y := int(yOffsetFactor*float64(width)) + int(math.Ceil(fontSize*dpi/72))
	d.Dot = fixed.Point26_6{
		X: (fixed.I(width) - d.MeasureString(letter)) / 2,
		Y: fixed.I(y),
	}
	d.DrawString(letter)

	b := bufio.NewWriter(out)
	encoder := png.Encoder{CompressionLevel: png.DefaultCompression}
	err := encoder.Encode(b, rgba)
	if err != nil {
		return err
	}
	err = b.Flush()
	if err != nil {
		return err
	}
	return nil
}

func pickForegroundColor(bgColor color.Color) color.Color {
	cWhite := contrastRatio(pickLighterColor(color.White, bgColor))

	// We prefer white text, this ratio was deemed good enough
	if cWhite > 1.5 {
		return color.White
	}
	return lightDark
}

func pickLighterColor(c1, c2 color.Color) (color.Color, color.Color) {
	_, _, v1 := RGBToHSV(c1)
	_, _, v2 := RGBToHSV(c2)

	if v1 >= v2 {
		return c1, c2
	}
	return c2, c1
}

// https://code.google.com/p/gorilla/source/browse/color/hsv.go?r=ef489f63418265a7249b1d53bdc358b09a4a2ea0
func RGBToHSV(c color.Color) (h, s, v float64) {
	r, g, b, _ := c.RGBA()
	fR := float64(r) / 255
	fG := float64(g) / 255
	fB := float64(b) / 255
	max := math.Max(math.Max(fR, fG), fB)
	min := math.Min(math.Min(fR, fG), fB)
	d := max - min
	s, v = 0, max
	if max > 0 {
		s = d / max
	}
	if max == min {
		// Achromatic.
		h = 0
	} else {
		// Chromatic.
		switch max {
		case fR:
			h = (fG - fB) / d
			if fG < fB {
				h += 6
			}
		case fG:
			h = (fB-fR)/d + 2
		case fB:
			h = (fR-fG)/d + 4
		}
		h /= 6
	}
	return
}

func contrastRatio(c1 color.Color, c2 color.Color) float64 {
	// http://www.w3.org/TR/2008/REC-WCAG20-20081211/#contrast-ratiodef

	l1 := relativeLuminance(c1)
	l2 := relativeLuminance(c2)

	return (l1 + 0.05) / (l2 + 0.05)
}

func relativeLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	r64 := foo(r)
	g64 := foo(g)
	b64 := foo(b)

	return 0.2126*r64 + 0.7152*g64 + 0.0722*b64
}

const shiftRGB = uint8(8)

func foo(col uint32) float64 {
	c := float64(uint8(col >> shiftRGB))
	c /= 255.0

	if c < 0.03928 {
		return c / 12.92
	}

	return math.Pow(((c + 0.055) / 1.055), 2.4)
}

var (
	errMalformedColorString = errors.New("Malformed hex color string")
)

func ColorFromHex(hex string) (*color.RGBA, error) {
	if len(hex) != 6 && len(hex) != 7 {
		return nil, errMalformedColorString
	}
	hex = strings.TrimPrefix(hex, "#")

	r, err := strconv.ParseInt(hex[0:2], 16, 16)
	if err != nil {
		return nil, errMalformedColorString
	}
	g, err := strconv.ParseInt(hex[2:4], 16, 16)
	if err != nil {
		return nil, errMalformedColorString
	}
	b, err := strconv.ParseInt(hex[4:6], 16, 16)
	if err != nil {
		return nil, errMalformedColorString
	}

	col := color.RGBA{uint8(r), uint8(g), uint8(b), 0xff}
	return &col, nil
}

func IconPath(letter string, size string, colr *color.RGBA) string {
	if letter == "" {
		letter = " "
	} else {
		letter = strings.ToUpper(letter)
	}

	if colr != nil {
		return fmt.Sprintf("/lettericons/%s-%s-%s.png", letter, size, colorfinder.ColorToHex(*colr))
	}
	return fmt.Sprintf("/lettericons/%s-%s.png", letter, size)
}

const defaultIconSize = 144
const maxIconSize = 1024

// path is like: lettericons/M-144-EFC25D.png
func ParseIconPath(fullpath string) (string, *color.RGBA, int) {
	_, filename := path.Split(fullpath)
	filename = strings.TrimSuffix(filename, ".png")
	params := strings.Split(filename, "-")
	if len(params) < 1 || len(params[0]) < 1 {
		return "", nil, -1
	}

	charParam := string(params[0][0])
	sizeParam := ""
	if len(params) >= 2 {
		sizeParam = params[1]
	}
	colorParam := ""
	if len(params) >= 3 {
		colorParam = params[2]
	}

	size, err := strconv.Atoi(sizeParam)
	if err != nil || size < 0 {
		size = defaultIconSize
	}
	if size > maxIconSize {
		size = maxIconSize
	}

	col, _ := ColorFromHex(colorParam)
	if col == nil {
		col = DefaultBackgroundColor
	}

	return charParam, col, size
}

func MainLetterFromURL(URL string) string {
	URL = strings.TrimSpace(URL)
	if !strings.HasPrefix(URL, "http") {
		URL = "http://" + URL
	}

	url, err := url.Parse(URL)
	if err != nil {
		return ""
	}

	host := url.Host
	hostSuffix, _ := publicsuffix.PublicSuffix(host)
	if hostSuffix != "" {
		host = strings.TrimSuffix(host, hostSuffix)
		host = strings.TrimSuffix(host, ".")
	}

	hostParts := strings.Split(host, ".")
	domain := hostParts[len(hostParts)-1]
	if len(domain) > 0 {
		return string(domain[0])
	} else if len(hostSuffix) > 0 {
		return string(hostSuffix[0])
	}

	return ""
}

var fnt *truetype.Font

var DefaultBackgroundColor *color.RGBA
var lightDark *color.RGBA

func init() {
	var err error
	fnt, err = truetype.Parse(fonts.OpenSansLightBytes())
	if err != nil {
		log.Println(err)
		return
	}

	DefaultBackgroundColor, _ = ColorFromHex("#909090")
	lightDark, _ = ColorFromHex("#505050")
}
