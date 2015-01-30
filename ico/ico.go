// Package ico register image.DecodeConfig support
// for the icon (container) format.
package ico

import (
	"encoding/binary"
	"image"
	"io"
)

type icondir struct {
	Reserved uint16
	Type     uint16
	Count    uint16
	Entries  []icondirEntry
}

type icondirEntry struct {
	Width    byte
	Height   byte
	Colors   byte
	Reserved byte
	Planes   uint16
	Bits     uint16
	Bytes    uint32
	Offset   uint32
}

func (dir *icondir) FindBestIcon() icondirEntry {
	best := icondirEntry{}
	for _, e := range dir.Entries {
		if (e.Width > best.Width) && (e.Height > best.Height) {
			best = e
		}
	}
	return best
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

func (e *icondirEntry) ColorCount() int {
	if e.Colors == 0 {
		return 256
	}
	return int(e.Colors)
}

// DecodeConfig returns just the dimensions of the largest image
// contained in the icon withou decoding the entire icon file.
func DecodeConfig(r io.Reader) (image.Config, error) {
	dir, err := ParseIco(r)
	if err != nil {
		return image.Config{}, err
	}

	best := dir.FindBestIcon()
	return image.Config{Width: int(best.Width), Height: int(best.Height)}, nil
}

const icoHeader = "\x00\x00\x01\x00"

func init() {
	image.RegisterFormat("ico", icoHeader, nil, DecodeConfig)
}
