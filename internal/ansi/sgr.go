package ansi

import (
	"image/color"
	"strconv"
)

// Color names may not match the standard (they don't match Wikipedia) in favor
// of making sense and save characters. For exmaple, "bright black" is just gray
// and the "bright" variants are default as they are more commonly used. This
// coincidentally also more closely matches Minecraft's naming scheme.
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

func Color(c color.Color) string {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return "\033[38;2;" + strconv.Itoa(int(n.R)) + ";" + strconv.Itoa(int(n.G)) + ";" + strconv.Itoa(int(n.B)) + "m"
}

func Bg(c color.Color) string {
	n := color.NRGBAModel.Convert(c).(color.NRGBA)
	return "\033[48;2;" + strconv.Itoa(int(n.R)) + ";" + strconv.Itoa(int(n.G)) + ";" + strconv.Itoa(int(n.B)) + "m"
}
