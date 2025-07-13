package print

import (
	"fmt"
	"image"
	"image/color"
	"minefetch/internal/ansi"
)

// The full block character (█) is not used in favor of setting the background
// color of a space character because some terminals (e.g. xterm out of the box)
// render block characters weird.
var blocks = [...]string{" ", "░", "▒", "▓", " "}

// Prints an image using Unicode shaded block characters (▓, ▒, ░) and ANSI
// 24-bit foreground and background color escape codes.
//
// Terminal fonts typically have an approximate aspect ratio of 1:2, i.e. pixels
// will be twice as tall as they are wide. The square parameter controls whether
// 2 characters are printed for every pixel, so as to make the pixels 1:1. This
// function does not perform any image scaling, so it is the responsibiliity of
// the caller to do so if desired.
//
// Transparency is supported by mapping the alpha channel to the 5 levels, from
// fully transparent to opaque, corresponding to a space character with no
// background color, shaded block characters and a space character with a
// colored background.
func BlockPrint(img image.Image, square bool) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			// https://pkg.go.dev/image/png#example-Decode
			level := c.A / 51 // 51 * 5 = 255
			if level == 5 {
				level--
			}
			var s string
			switch {
			case level == 0:
				s = ansi.ResetBg + " "
			case level < uint8(len(blocks))-1:
				s = ansi.Color(c) + blocks[level]
			default:
				s = ansi.Bg(c) + " "
			}
			if square {
				s += blocks[level]
			}
			fmt.Print(s)
		}
		// It's blockPrint, not blockPrintln
		if y != b.Max.Y-1 {
			fmt.Print(ansi.ResetBg + "\n")
		}
	}
	fmt.Print(ansi.Reset)
}

// Prints an iamge using a combination of Unicde upper and lower half block
// characters (▀, ▄) and ANSI 24-bit color foreground and background escape
// codes.
//
// Terminal fonts typically have an approximate aspect ratio of 1:2, i.e. pixels
// will be twice as tall as they are wide, so this method should reverse that
// effect and print close to a square aspect ratio.
//
// There are no shaded variants of these half block characters, so transparency
// support is limited to a threshold value that determines whether the half part
// of the the character corresponding to the pixel is drawn or not, i.e. 2
// levels of transparency.
//
// This method was inspired by [catimg] and [pixterm].
//
// [catimg]: https://github.com/posva/catimg
// [pixterm]: https://github.com/eliukblau/pixterm
func HalfPrint(img image.Image, thresh uint8) {
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y += 2 {
		for x := b.Min.X; x < b.Max.X; x++ {
			c0 := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			c1 := color.NRGBAModel.Convert(img.At(x, y+1)).(color.NRGBA)
			// Background color is only used if both pixels satisfy the alpha
			// threshold
			if c0.A >= thresh && c1.A >= thresh {
				fmt.Print(ansi.Bg(c0) + ansi.Color(c1) + "▄")
			} else if c0.A >= thresh {
				fmt.Print(ansi.ResetBg + ansi.Color(c0) + "▀")
			} else if c1.A >= thresh {
				fmt.Print(ansi.ResetBg + ansi.Color(c1) + "▄")
			} else {
				fmt.Print(ansi.ResetBg + " ")
			}
		}
		if y+2 < b.Max.Y {
			fmt.Print(ansi.Reset + "\n")
		}
	}
	fmt.Print(ansi.Reset)
}
