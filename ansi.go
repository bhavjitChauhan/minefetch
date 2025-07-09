package main

import (
	"image/color"
	"strconv"
)

// https://en.wikipedia.org/wiki/ANSI_escape_code#Select_Graphic_Rendition_parameters
const (
	reset     = "\033[0m"
	bold      = "\033[1m"
	italic    = "\033[3m"
	underline = "\033[4m"
	invert    = "\033[7m"
	strike    = "\033[9m"
)

func curUp(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "A"
}

func curDown(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "B"
}

func curFwd(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "C"
}

func curBack(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "D"
}

func trueColor(c color.NRGBA) string {
	return "\033[38;2;" + strconv.Itoa(int(c.R)) + ";" + strconv.Itoa(int(c.G)) + ";" + strconv.Itoa(int(c.B)) + "m"
}
