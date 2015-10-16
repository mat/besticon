// Package ico register image.DecodeConfig support
// for the icon (container) format.
package ico

import (
	"encoding/binary"
	"errors"
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

func (e *icondirEntry) ColorCount() int {
	if e.Colors == 0 {
		return 256
	}
	return int(e.Colors)
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

const icoHeader = "\x00\x00\x01\x00"

func init() {
	image.RegisterFormat("ico", icoHeader, nil, DecodeConfig)
}
