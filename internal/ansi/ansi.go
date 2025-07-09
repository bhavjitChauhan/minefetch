package ansi

import (
	"image/color"
	"strconv"
)

// https://en.wikipedia.org/wiki/ANSI_escape_code#Select_Graphic_Rendition_parameters
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Italic    = "\033[3m"
	Underline = "\033[4m"
	Invert    = "\033[7m"
	Strike    = "\033[9m"
)

func Up(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "A"
}

func Down(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "B"
}

func Fwd(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "C"
}

func Back(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "D"
}

func Color(c color.NRGBA) string {
	return "\033[38;2;" + strconv.Itoa(int(c.R)) + ";" + strconv.Itoa(int(c.G)) + ";" + strconv.Itoa(int(c.B)) + "m"
}
