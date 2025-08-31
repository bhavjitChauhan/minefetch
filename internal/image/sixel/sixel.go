/*
Package sixel implements the [Sixel] image format.

Although it is the most widely-supported method to display images in terminals,
the implementation and behavior is not standardized.
For example, some terminals support transparent backgrounds,
while others use the first color in the palette as the background.
Prefer other terminal images formats, like those used by [iTerm2] or [kitty].

[Sixel]: https://en.wikipedia.org/wiki/Sixel
[iTerm2]: https://iterm2.com/documentation-images.html
[kitty]: https://sw.kovidgoyal.net/kitty/graphics-protocol
*/
package sixel

// https://www.vt100.net/docs/vt3xx-gp/chapter14.html
// https://shuford.invisible-island.net/all_about_sixels.txt
// https://github.com/saitoha/libsixel

import (
	"bytes"
	"cmp"
	"image"
	"image/color"
	"image/draw"
	"io"
	"strconv"

	"github.com/bhavjitChauhan/minefetch/internal/image/quant"
)

type Options struct {
	// Fallback is the background color for alpha compositing used by some terminals (e.g. Konsole, Rio).
	Fallback color.Color
	// Thresh is the run-length encoding threshold (RLE).
	// It is the minimum number of consecutive occurrences of the same byte that will be compressed.
	// A value of 0 will disable RLE.
	// A value of 1 will encode every byte; this should only be used for experimentation.
	Thresh uint
}

// Encode writes the Image m to w in the [Sixel] format.
//
// The first and second colors in the palette are reserved for Options.Fallback and color.Transparent.
//
// [Sixel]: https://en.wikipedia.org/wiki/Sixel
func Encode(w io.Writer, m image.Image, o *Options) error {
	p := make(color.Palette, 0, 256)
	fallback := color.Color(color.Black)
	thresh := uint(4)
	if o != nil {
		if o.Fallback != nil {
			fallback = o.Fallback
		}
		thresh = o.Thresh
	}
	p = append(p, fallback)
	p = append(p, color.Transparent)
	// TODO: skip quantization if already paletted and palette size < 255 (bg + transparent?)
	p = quant.MedianCut.Quantize(p, m)
	bounds := m.Bounds()
	pm := image.NewPaletted(bounds, p)
	// TODO: this is slow
	draw.FloydSteinberg.Draw(pm, bounds, m, image.Point{})
	b := &bytes.Buffer{}
	_, err := b.WriteString("\033Pq")
	if err != nil {
		return err
	}
	err = writeAttributes(b, 1, 1, bounds.Dx(), bounds.Dy())
	if err != nil {
		return err
	}
	err = writePalette(b, p)
	if err != nil {
		return err
	}
	err = writeData(b, pm, thresh)
	if err != nil {
		return err
	}
	b.WriteString("\033\\")
	_, err = b.WriteTo(w)
	return err
}

// Some terminals that don't support transparent background use width and height to cut off the image.
// Pixel sizes other than 1 by 1 are not widely supported.
//
// https://www.vt100.net/docs/vt3xx-gp/chapter14.html#:~:text=Raster%20Attributes%20(%22)
func writeAttributes(b *bytes.Buffer, pixelWidth, pixelHeight uint, width, height int) error {
	err1 := b.WriteByte('"')
	_, err2 := b.WriteString(strconv.FormatUint(uint64(pixelHeight), 10))
	err3 := b.WriteByte(';')
	_, err4 := b.WriteString(strconv.FormatUint(uint64(pixelWidth), 10))
	err5 := b.WriteByte(';')
	_, err6 := b.WriteString(strconv.Itoa(width))
	err7 := b.WriteByte(';')
	_, err8 := b.WriteString(strconv.Itoa(height))
	return cmp.Or(err1, err2, err3, err4, err5, err6, err7, err8)
}

// https://www.vt100.net/docs/vt3xx-gp/chapter14.html#:~:text=Color%20Introducer%20(%23)
func writePalette(b *bytes.Buffer, p color.Palette) error {
	for i, c := range p {
		c := color.NRGBAModel.Convert(c).(color.NRGBA)
		err1 := b.WriteByte('#')
		_, err2 := b.WriteString(strconv.Itoa(i))
		_, err3 := b.WriteString(";2;")
		_, err4 := b.WriteString(strconv.FormatUint(uint64(c.R)*100/255, 10))
		err5 := b.WriteByte(';')
		_, err6 := b.WriteString(strconv.FormatUint(uint64(c.G)*100/255, 10))
		err7 := b.WriteByte(';')
		_, err8 := b.WriteString(strconv.FormatUint(uint64(c.B)*100/255, 10))
		if err := cmp.Or(err1, err2, err3, err4, err5, err6, err7, err8); err != nil {
			return err
		}
	}
	return nil
}

func writeData(b *bytes.Buffer, m *image.Paletted, thresh uint) error {
	bounds := m.Bounds()
	prevIndex := uint8(0)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 6 {
		colors := make(map[uint8][]byte)
		maxDy := 6
		if y+6 > bounds.Max.Y {
			maxDy = bounds.Max.Y - y
		}
		for dy := range maxDy {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				index := m.ColorIndexAt(x, y+dy)
				// Transparent
				if index == 1 {
					continue
				}
				if _, ok := colors[index]; !ok {
					colors[index] = make([]byte, bounds.Dx())
				}
				colors[index][x] |= 1 << dy
			}
		}
		i := 0
		for index, bytes := range colors {
			if i != 0 || index != prevIndex {
				err1 := b.WriteByte('#')
				_, err2 := b.WriteString(strconv.FormatUint(uint64(index), 10))
				if err := cmp.Or(err1, err2); err != nil {
					return err
				}
			}
			run := 1
			for j := 1; j <= len(bytes); j++ {
				// TODO: cut off runs at 255 for some terminals?
				if j != len(bytes) && bytes[j] == bytes[j-1] {
					run++
					continue
				}
				if thresh != 0 && run >= int(thresh) {
					err1 := b.WriteByte('!')
					_, err2 := b.WriteString(strconv.Itoa(run))
					if err := cmp.Or(err1, err2); err != nil {
						return err
					}
					run = 1
				}
				for range run {
					err := b.WriteByte('?' + bytes[j-1])
					if err != nil {
						return err
					}
				}
				run = 1
			}
			if i < len(colors)-1 {
				err := b.WriteByte('$')
				if err != nil {
					return err
				}
			} else {
				prevIndex = index
			}
			i++
		}
		err := b.WriteByte('-')
		if err != nil {
			return err
		}
	}
	return nil
}
