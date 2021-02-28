package lettericon

import (
	"bufio"
	"bytes"
	"encoding/xml"
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
	"path/filepath"
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

func RenderPNG(letter string, bgColor color.Color, width int, out io.Writer) error {
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
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
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

func IconPath(letter string, size string, colr *color.RGBA, format string) string {
	var parts []string

	// letter
	if letter == "" {
		letter = " "
	} else {
		letter = strings.ToUpper(letter)
	}
	parts = append(parts, letter)

	// size (maybe)
	if format == "png" {
		parts = append(parts, size)
	}

	// colr (maybe)
	if colr != nil {
		parts = append(parts, colorfinder.ColorToHex(*colr))
	}

	return fmt.Sprintf("/lettericons/%s.%s", strings.Join(parts, "-"), format)
}

const defaultIconSize = 144

// TODO: Sync with besticon.MaxIconSize ?
const maxIconSize = 256

// path is like: lettericons/M-144-EFC25D.png
func ParseIconPath(fullpath string) (string, *color.RGBA, int, string) {
	fullpath = percentDecode(fullpath)

	_, filename := path.Split(fullpath)

	// format
	format := filepath.Ext(filename)
	if !(format == ".png" || format == ".svg") {
		return "", nil, -1, ""
	}
	filename = strings.TrimSuffix(filename, format)
	format = format[1:] // remove period

	// now we parse each of the params, delimited by "-"
	params := strings.Split(filename, "-")
	if len(params) == 0 {
		return "", nil, -1, ""
	}
	for _, s := range params {
		if len(s) == 0 {
			return "", nil, -1, ""
		}
	}

	var letter string
	var size int
	var col *color.RGBA

	// letter
	letter, params = firstRune(params[0]), params[1:]

	// size (only png)
	if format == "png" && len(params) > 0 {
		size, _ = strconv.Atoi(params[0])
		params = params[1:]
	}
	if size < 1 {
		size = defaultIconSize
	}
	if size > maxIconSize {
		size = maxIconSize
	}

	// color
	if len(params) > 0 {
		col, _ = ColorFromHex(params[0])
		params = params[1:]
	}
	if col == nil {
		col = DefaultBackgroundColor
	}

	// extra stuff at the end? error
	if len(params) > 0 {
		return "", nil, -1, ""
	}

	return letter, col, size, format
}

func MainLetterFromURL(URL string) string {
	URL = strings.TrimSpace(URL)
	if !strings.HasPrefix(URL, "http:") && !strings.HasPrefix(URL, "https:") {
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
		return firstRune(domain)
	} else if len(hostSuffix) > 0 {
		return string(hostSuffix[0])
	}

	return ""
}

func firstRune(str string) string {
	for _, runeValue := range str {
		return fmt.Sprintf("%c", runeValue)
	}
	return ""
}

func percentDecode(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	u, err := url.ParseRequestURI(p)

	if err != nil {
		return p
	}
	return u.Path
}

const svgTemplate = `
<svg width="100" height="100" xmlns="http://www.w3.org/2000/svg">
  <rect x="0" y="0" width="100" height="100" fill="$BG_COLOR"/>
  <text x="50%" y="50%" dy="0.10em" font-family="Helvetica Neue, Helvetica, sans-serif" font-size="75" dominant-baseline="middle" text-anchor="middle" fill="$FG_COLOR">$LETTER</text>
</svg>
`

// RenderSVG writes an SVG lettericon for this letter and color
func RenderSVG(letter string, bgColor color.Color, out io.Writer) error {
	// xml escape letter
	var buf bytes.Buffer
	err := xml.EscapeText(&buf, []byte(letter))
	if err != nil {
		return err
	}

	// vars
	vars := map[string]string{
		"$BG_COLOR": ColorToHex(bgColor),
		"$FG_COLOR": ColorToHex(pickForegroundColor(bgColor)),
		"$LETTER":   buf.String(),
	}

	// render SVG by replacing vars in template
	svg := strings.TrimSpace(svgTemplate) + "\n"
	for k, v := range vars {
		svg = strings.ReplaceAll(svg, k, v)
	}

	_, err = io.WriteString(out, svg)
	return err
}

// ColorToHex returns the #rrggbb hex string for a color
func ColorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", r&0xff, g&0xff, b&0xff)
}

var fnt *truetype.Font

var DefaultBackgroundColor *color.RGBA
var lightDark *color.RGBA

func init() {
	var err error
	fnt, err = truetype.Parse(fonts.NotoSansRegularBytes())
	if err != nil {
		log.Println(err)
		return
	}

	DefaultBackgroundColor, _ = ColorFromHex("#909090")
	lightDark, _ = ColorFromHex("#505050")
}
