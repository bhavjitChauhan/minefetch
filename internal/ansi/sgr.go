package ansi

import (
	"fmt"
	"image/color"
	"strconv"
)

// [Select Graphic Rendition parameters] control foreground and background style and color.
//
// Color names may not match the standard (they don't match Wikipedia) in favor of being more accurate and save characters.
// For example, "bright black" is just gray and the "bright" variants are default as they are more commonly used.
// This coincidentally also more closely matches Minecraft's naming scheme.
//
// These variables will be overwritten to empty string values by NoColor.
//
// [Select Graphic Rendition parameters]: https://en.wikipedia.org/wiki/ANSI_escape_code#Select_Graphic_Rendition_parameters
var (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Invert    = "\033[7m"
	Strike    = "\033[9m"
	ResetBg   = "\033[49m"

	Black       = "\033[30m"
	DarkRed     = "\033[31m"
	DarkGreen   = "\033[32m"
	DarkYellow  = "\033[33m"
	DarkBlue    = "\033[34m"
	DarkMagenta = "\033[35m"
	DarkCyan    = "\033[36m"
	LightGray   = "\033[37m"

	Gray    = "\033[90m"
	Red     = "\033[91m"
	Green   = "\033[92m"
	Yellow  = "\033[93m"
	Blue    = "\033[94m"
	Magenta = "\033[95m"
	Cyan    = "\033[96m"
	White   = "\033[97m"
)

// Color returns the sequence to set the foreground color to c.
// It respects ColorSupport and will convert c to the closest supported color, if needed.
func Color(c color.Color) string {
	switch ColorSupport {
	case NoColorSupport:
		return ""
	case Color16Support:
		return fg16(c)
	case Color256Support:
		return fg256(c)
	case TrueColorSupport:
		return fg(c)
	}
	panic(fmt.Sprint("invalid color support: ", ColorSupport))
}

// Bg returns the sequence to set the background color to c.
// It respects ColorSupport and will convert c to the closest supported color, if needed.
func Bg(c color.Color) string {
	switch ColorSupport {
	case NoColorSupport:
		return ""
	case Color16Support:
		return bg16(c)
	case Color256Support:
		return bg256(c)
	case TrueColorSupport:
		return bg(c)
	}
	panic(fmt.Sprint("invalid color support: ", ColorSupport))
}

func fg16(c color.Color) string {
	c = color16.Convert(c)
	code := uint8(color16.Index(c))
	if code < 8 {
		code += 30
	} else {
		code += 90 - 8
	}
	return "\033[" + strconv.Itoa(int(code)) + "m"
}

func bg16(c color.Color) string {
	c = color16.Convert(c)
	code := uint8(color16.Index(c))
	if code < 8 {
		code += 40
	} else {
		code += 100 - 8
	}
	return "\033[" + strconv.Itoa(int(code)) + "m"
}

func code256(c color.Color) uint8 {
	c = color256.Convert(c)
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	var r, g, b = n.R, n.G, n.B
	if r == g && g == b && r != 0 && r != 255 {
		return (r-8)/10 + 232
	}
	if r != 0 {
		r = (r - 55) / 40
	}
	if g != 0 {
		g = (g - 55) / 40
	}
	if b != 0 {
		b = (b - 55) / 40
	}
	return 16 + (r * 36) + (g * 6) + b
}

func fg256(c color.Color) string {
	code := code256(c)
	return "\033[38;5;" + strconv.Itoa(int(code)) + "m"
}

func bg256(c color.Color) string {
	code := code256(c)
	return "\033[48;5;" + strconv.Itoa(int(code)) + "m"
}

func fg(c color.Color) string {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return "\033[38;2;" + strconv.Itoa(int(n.R)) + ";" + strconv.Itoa(int(n.G)) + ";" + strconv.Itoa(int(n.B)) + "m"
}

func bg(c color.Color) string {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return "\033[48;2;" + strconv.Itoa(int(n.R)) + ";" + strconv.Itoa(int(n.G)) + ";" + strconv.Itoa(int(n.B)) + "m"
}
