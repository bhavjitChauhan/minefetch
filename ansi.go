package main

import (
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

	black   = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"

	brightBlack   = "\033[90m"
	brightRed     = "\033[91m"
	brightGreen   = "\033[92m"
	brightYellow  = "\033[93m"
	brightBlue    = "\033[94m"
	brightMagenta = "\033[95m"
	brightCyan    = "\033[96m"
	brightWhite   = "\033[97m"
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

func trueColor(r, g, b uint8) string {
	return "\033[38;2;" + strconv.Itoa(int(r)) + ";" + strconv.Itoa(int(g)) + ";" + strconv.Itoa(int(b)) + "m"
}
