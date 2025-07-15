// Like image.Config, but provides more information specific to the PNG format.
package pngconfig

import (
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type BitDepth byte

// https://www.w3.org/TR/png/#11IHDR
const (
	BitDepth1  BitDepth = 1
	BitDepth2  BitDepth = 1
	BitDepth4  BitDepth = 1
	BitDepth8  BitDepth = 1
	BitDepth16 BitDepth = 1
)

type ColorType byte

// https://www.w3.org/TR/png/#3colourType
const (
	ColorTypeGray    ColorType = 0
	ColorTypeRGB     ColorType = 2
	ColorTypeIndexed ColorType = 3
	ColorTypeGrayA   ColorType = 4
	ColorTypeRGBA    ColorType = 6
)

// https://www.w3.org/TR/png/#11IHDR
type Config struct {
	Width, Height       uint32
	BitDepth            BitDepth
	ColorType           ColorType
	Compression, Filter byte
	Interlaced          bool
}

func DecodeConfig(r io.Reader) (Config, error) {
	var c Config

	// Discard file header + chunk metadata
	n, err := io.CopyN(io.Discard, r, 16)
	if err != nil {
		return c, err
	}
	if n != 16 {
		return c, errors.New(fmt.Sprint("expected 16 bytes, got ", n))
	}

	err1 := binary.Read(r, binary.BigEndian, &c.Width)
	err2 := binary.Read(r, binary.BigEndian, &c.Height)
	err3 := binary.Read(r, binary.BigEndian, &c.BitDepth)
	err4 := binary.Read(r, binary.BigEndian, &c.ColorType)
	err5 := binary.Read(r, binary.BigEndian, &c.Compression)
	err6 := binary.Read(r, binary.BigEndian, &c.Filter)
	err7 := binary.Read(r, binary.BigEndian, &c.Interlaced)

	return c, cmp.Or(err1, err2, err3, err4, err5, err6, err7)
}
