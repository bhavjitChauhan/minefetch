package ansi

import (
	"image/color"
	"strconv"
)

// [Select Graphic Rendition parameters] control foreground and background style and color.
//
// Color names may not match the standard (they don't match Wikipedia) in favor of being more accurate and save characters.
// For example, "bright black" is just gray and the "bright" variants are default as they are more commonly used.
// This coincidentally also more closely matches Minecraft's naming scheme.
//
// [Select Graphic Rendition parameters]: https://en.wikipedia.org/wiki/ANSI_escape_code#Select_Graphic_Rendition_parameters
const (
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
func Color(c color.Color) string {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return "\033[38;2;" + strconv.Itoa(int(n.R)) + ";" + strconv.Itoa(int(n.G)) + ";" + strconv.Itoa(int(n.B)) + "m"
}

// Bg returns the sequence to set the background color to c.
func Bg(c color.Color) string {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return "\033[48;2;" + strconv.Itoa(int(n.R)) + ";" + strconv.Itoa(int(n.G)) + ";" + strconv.Itoa(int(n.B)) + "m"
}
