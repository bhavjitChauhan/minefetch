// Package term provides various facilities to work with terminal emulators.
//
// Only Linux, macOS and Windows are supported.
package term

// https://invisible-island.net/xterm/ctlseqs/ctlseqs.html#h3-Miscellaneous

import (
	"bytes"
	"fmt"
	"strconv"
)

// QuerySize gets the terminal size in pixels using an xterm code.
func QuerySize() (width, height uint, err error) {
	res, err := query("\033[18t")
	if err != nil {
		return
	}
	buf := bytes.NewBuffer([]byte(res))
	s, err := buf.ReadString(';')
	if err != nil {
		return
	}
	if s != "\033[8;" {
		err = fmt.Errorf("unexpected response: %v", res)
		return
	}
	s, err = buf.ReadString(';')
	if err != nil {
		return
	}
	huint64, err := strconv.ParseUint(s[:len(s)-1], 10, 0)
	if err != nil {
		return
	}
	s, err = buf.ReadString('t')
	if err != nil {
		return
	}
	wuint64, err := strconv.ParseUint(s[:len(s)-1], 10, 0)
	if err != nil {
		return
	}

	width = uint(wuint64)
	height = uint(huint64)
	return
}

// QueryCellSize gets the terminal cell size in pixels using xterm codes.
// Since some terminals do not report cell sizes in pixels directly, two queries are used.
func QueryCellSize() (width, height uint, err error) {
	res, err := query("\033[14t", "\033[18t")
	if err != nil {
		return
	}
	buf := bytes.NewBuffer([]byte(res))
	s, err := buf.ReadString(';')
	if err != nil {
		return
	}
	if s != "\033[4;" {
		err = fmt.Errorf("unexpected response: %v", res)
		return
	}
	s, err = buf.ReadString(';')
	if err != nil {
		return
	}
	winHeight, err := strconv.ParseUint(s[:len(s)-1], 10, 0)
	if err != nil {
		return
	}
	s, err = buf.ReadString('t')
	if err != nil {
		return
	}
	winWidth, err := strconv.ParseUint(s[:len(s)-1], 10, 0)
	if err != nil {
		return
	}
	s, err = buf.ReadString(';')
	if err != nil {
		return
	}
	if s != "\033[8;" {
		err = fmt.Errorf("unexpected response: %v", res)
		return
	}
	s, err = buf.ReadString(';')
	if err != nil {
		return
	}
	winHeightCells, err := strconv.ParseUint(s[:len(s)-1], 10, 0)
	if err != nil {
		return
	}
	s, err = buf.ReadString('t')
	if err != nil {
		return
	}
	winWidthCells, err := strconv.ParseUint(s[:len(s)-1], 10, 0)
	if err != nil {
		return
	}

	width = uint(winWidth / winWidthCells)
	height = uint(winHeight / winHeightCells)
	return
}
