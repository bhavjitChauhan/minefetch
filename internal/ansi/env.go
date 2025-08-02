package ansi

import (
	"os"
	"strings"
)

const (
	NoColorSupport   = iota
	Color16Support   // 4-bit color
	Color256Support  // 8-bit color
	TrueColorSupport // 24-bit color
)

// ColorSupport is the detected terminal color support level.
//
// Currently, support is detected exclusively through environment variables.
// In addition to program variables, FORCE_COLOR can be specified to specify a specific support level:
//   - 0 (no color)
//   - 1 (8-bit color)
//   - 2 (16-bit color)
//   - 3 (24-bit true color)
//
// The NO_COLOR, CLICOLOR_FORCE and CLICOLOR variables are also supported.
var ColorSupport = NoColorSupport

// https://xkcd.com/927
func init() {
	// https://force-color.org
	forceColor := os.Getenv("FORCE_COLOR")
	if forceColor != "" {
		// https://github.com/chalk/supports-color/blob/ae809ecabd5965d0685e7fc121efe98c47ad8724/index.js#L33
		switch forceColor {
		case "0", "false":
			ColorSupport = NoColorSupport
		case "2":
			ColorSupport = Color256Support
		case "3":
			ColorSupport = TrueColorSupport
		// case "1", "true":
		default:
			ColorSupport = Color16Support
		}
		return
	}
	// The *Support functions do not cascade. That is, a terminal with
	// trueColorSupport true may have color256Support false.
	if noColorSupport() {
		ColorSupport = NoColorSupport
		return
	}
	if trueColorSupport() {
		ColorSupport = TrueColorSupport
		return
	}
	if color256Support() {
		ColorSupport = Color256Support
		return
	}
	if color16Support() {
		ColorSupport = Color16Support
		return
	}
}

func noColorSupport() bool {
	// https://no-color.org
	noColor := os.Getenv("NO_COLOR")
	// https://github.com/cli/go-gh/blob/a08820a13f257d6c5b4cb86d37db559ec6d14577/pkg/term/env.go#L161
	cliColorForce := os.Getenv("CLICOLOR_FORCE")
	cliColor := os.Getenv("CLICOLOR")
	return noColor != "" || cliColorForce == "0" || cliColor == "0"
}

func trueColorSupport() bool {
	term := os.Getenv("TERM")
	// https://github.com/termstandard/colors
	colorTerm := os.Getenv("COLORTERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	// Not ideal; https://github.com/Textualize/rich/issues/140
	wtSession := os.Getenv("WT_SESSION")
	return strings.Contains(term, "24bit") ||
		strings.Contains(term, "truecolor") ||
		strings.Contains(colorTerm, "24bit") ||
		strings.Contains(colorTerm, "truecolor") ||
		termProgram == "mintty" ||
		termProgram == "waveterm" ||
		wtSession != ""
}

func color256Support() bool {
	term := os.Getenv("TERM")
	// https://github.com/cli/go-gh/blob/a08820a13f257d6c5b4cb86d37db559ec6d14577/pkg/term/env.go#L170
	colorTerm := os.Getenv("COLORTERM")
	return strings.Contains(term, "256") || strings.Contains(colorTerm, "256")
}

func color16Support() bool {
	// https://bixense.com/clicolors
	cliColorForce := os.Getenv("CLICOLOR_FORCE")
	cliColor := os.Getenv("CLICOLOR")
	// https://github.com/adoxa/ansicon
	ansicon := os.Getenv("ANSICON")
	return cliColorForce != "" || cliColor != "" || ansicon != ""
}

// NoColor overwrites the SGR variables to empty string values.
// This reduces effort needed to integrate no-color support.
//
// This action is irreversible!
// As such, it is not called by the package in case the application wishes to override the color support.
// The application should call it if ColorSupport == NoColorSupport is true after any configuration parsing.
func NoColor() {
	ColorSupport = NoColorSupport
	Reset = ""
	Bold = ""
	Italic = ""
	Underline = ""
	Invert = ""
	Strike = ""
	ResetBg = ""
	Black = ""
	DarkRed = ""
	DarkGreen = ""
	DarkYellow = ""
	DarkBlue = ""
	DarkMagenta = ""
	DarkCyan = ""
	LightGray = ""
	Gray = ""
	Red = ""
	Green = ""
	Yellow = ""
	Blue = ""
	Magenta = ""
	Cyan = ""
	White = ""
}
