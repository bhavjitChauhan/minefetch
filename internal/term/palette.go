package term

import (
	"image/color"
)

// color16 is a 16-color palette that is used for 4-bit terminal colors.
// These colors used exclusively for palette conversion.
// They are likely inaccurate as they are implementation-defined.
//
// https://en.wikipedia.org/wiki/ANSI_escape_code#3-bit_and_4-bit
var color16 = color.Palette{
	color.NRGBA{0, 0, 0, 255},       // Black
	color.NRGBA{128, 0, 0, 255},     // DarkRed
	color.NRGBA{0, 128, 0, 255},     // DarkGreen
	color.NRGBA{128, 128, 0, 255},   // DarkYellow
	color.NRGBA{0, 0, 128, 255},     // DarkBlue
	color.NRGBA{128, 0, 128, 255},   // DarkMagenta
	color.NRGBA{0, 128, 128, 255},   // DarkCyan
	color.NRGBA{192, 192, 192, 255}, // LightGray
	color.NRGBA{128, 128, 128, 255}, // Gray
	color.NRGBA{255, 0, 0, 255},     // Red
	color.NRGBA{0, 255, 0, 255},     // Green
	color.NRGBA{255, 255, 0, 255},   // Yellow
	color.NRGBA{0, 0, 255, 255},     // Blue
	color.NRGBA{255, 0, 255, 255},   // Magenta
	color.NRGBA{0, 255, 255, 255},   // Cyan
	color.NRGBA{255, 255, 255, 255}, // White
}

// color256 is an 256-color palette that is used for 8-bit terminal colors.
// Although these colors are also not standardized, there is necessarily less deviation.
//
// https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
var color256 color.Palette

func init() {
	color256 = append(color256, color16...)
	for red := range uint8(6) {
		for green := range uint8(6) {
			for blue := range uint8(6) {
				var r, g, b uint8
				if red != 0 {
					r = red*40 + 55
				}
				if green != 0 {
					g = green*40 + 55
				}
				if blue != 0 {
					b = blue*40 + 55
				}
				color256 = append(color256, color.NRGBA{r, g, b, 255})
			}
		}
	}
	for gray := range uint8(24) {
		g := gray*10 + 8
		color256 = append(color256, color.NRGBA{g, g, g, 255})
	}
}
