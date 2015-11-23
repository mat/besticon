// Package ico register image.DecodeConfig support
// for the icon (container) format.
package ico

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"os"

	//asdf
	_ "image/gif"
	_ "image/png"
)

type icondir struct {
	Reserved uint16
	Type     uint16
	Count    uint16
	Entries  []icondirEntry
}

type icondirEntry struct {
	Width        byte
	Height       byte
	PaletteCount byte
	Reserved     byte
	ColorPlanes  uint16
	BitsPerPixel uint16
	Size         uint32
	Offset       uint32
}

func (dir *icondir) FindBestIcon() *icondirEntry {
	if len(dir.Entries) == 0 {
		return nil
	}

	best := dir.Entries[0]
	for _, e := range dir.Entries {
		if (e.width() > best.width()) && (e.height() > best.height()) {
			best = e
		}
	}
	return &best
}

// ParseIco parses the icon and returns meta information for the icons as icondir.
func ParseIco(r io.Reader) (*icondir, error) {
	dir := icondir{}

	var err error
	err = binary.Read(r, binary.LittleEndian, &dir.Reserved)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &dir.Type)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &dir.Count)
	if err != nil {
		return nil, err
	}

	for i := uint16(0); i < dir.Count; i++ {
		entry := icondirEntry{}
		err := parseIcondirEntry(r, &entry)
		if err != nil {
			return nil, err
		}
		dir.Entries = append(dir.Entries, entry)
	}

	return &dir, err
}

func parseIcondirEntry(r io.Reader, e *icondirEntry) error {
	err := binary.Read(r, binary.LittleEndian, e)
	if err != nil {
		return err
	}

	return nil
}

type dibHeader struct {
	dibHeaderSize uint32
	width         uint32
	height        uint32
}

func (e *icondirEntry) ColorCount() int {
	if e.PaletteCount == 0 {
		return 256
	}
	return int(e.PaletteCount)
}

func (e *icondirEntry) width() int {
	if e.Width == 0 {
		return 256
	}
	return int(e.Width)
}

func (e *icondirEntry) height() int {
	if e.Height == 0 {
		return 256
	}
	return int(e.Height)
}

// DecodeConfig returns just the dimensions of the largest image
// contained in the icon withou decoding the entire icon file.
func DecodeConfig(r io.Reader) (image.Config, error) {
	dir, err := ParseIco(r)
	if err != nil {
		return image.Config{}, err
	}

	best := dir.FindBestIcon()
	if best == nil {
		return image.Config{}, errors.New("ico file does not contain any icons")
	}
	return image.Config{Width: best.width(), Height: best.height()}, nil
}

type header struct {
	sigBM     [2]byte
	fileSize  uint32
	resverved [2]uint16
	pixOffset uint32
	// dibHeaderSize   uint32
	// width           uint32
	// height          uint32
	// colorPlane      uint16
	// bpp             uint16
	// compression     uint32
	// imageSize       uint32
	// xPixelsPerMeter uint32
	// yPixelsPerMeter uint32
	// colorUse        uint32
	// colorImportant  uint32
}

func decodeImage(r io.Reader) (image.Image, error) {
	dir, err := ParseIco(r)
	if err != nil {
		return nil, err
	}

	best := dir.FindBestIcon()
	if best == nil {
		return nil, errors.New("ico file does not contain any icons")
	}

	startOffset := best.Offset
	endOffset := startOffset + best.Size
	fullIcoBytes, err := ioutil.ReadAll(r)
	singleIcoBytes := fullIcoBytes[startOffset:(endOffset)]

	h := &header{
		sigBM:     [2]byte{'B', 'M'},
		fileSize:  14 + best.Size,
		pixOffset: 14,
		// dibHeaderSize: 40,
		// width:         uint32(best.Width),
		// height:        uint32(best.Height),
		// colorPlane:    1,
	}

	buf := new(bytes.Buffer)
	if err = binary.Write(buf, binary.LittleEndian, h); err != nil {
		return nil, err
	}
	buf.Write(singleIcoBytes)

	// step = (3*d.X + 3) &^ 3
	// h.imageSize = uint32(d.Y * step)
	// h.fileSize += h.imageSize

	fmt.Println("len(bytes)", buf.Len())

	f, err := os.Create("/tmp/pic.bmp")
	if err != nil {
		return nil, err
	}
	f.Write(buf.Bytes())
	f.Sync()
	defer f.Close()

	// err = ioutil.WriteFile("/tmp/pic.bmp", buf.Bytes(), 0644)
	// if err != nil {
	// 	return nil, err
	// }

	img, format, err := image.Decode(bytes.NewReader(buf.Bytes()))
	fmt.Println("err:", err)
	fmt.Println("format:", format)

	return img, err
}

const icoHeader = "\x00\x00\x01\x00"

func init() {
	image.RegisterFormat("ico", icoHeader, decodeImage, DecodeConfig)
}
