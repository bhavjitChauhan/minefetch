//go:build windows

package term

import (
	"syscall"
	"unsafe"
)

// https://learn.microsoft.com/en-us/windows/console/coord-str
type coord struct {
	x int16
	y int16
}

// https://learn.microsoft.com/en-us/windows/console/small-rect-str
type smallRect struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

// https://learn.microsoft.com/en-us/windows/console/console-screen-buffer-info-str
type consoleScreenBufferInfo struct {
	size              coord
	cursorPosition    coord
	attributes        uint16
	window            smallRect
	maximumWindowSize coord
}

const (
	_ENABLE_ECHO_INPUT             = 0x0004
	_ENABLE_PROCESSED_INPUT        = 0x0001
	_ENABLE_LINE_INPUT             = 0x0002
	_ENABLE_VIRTUAL_TERMINAL_INPUT = 0x0200
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
var procSetConsoleMode = kernel32.NewProc("SetConsoleMode")

func size() (width, height uint, err error) {
	var info consoleScreenBufferInfo
	_, _, errno := syscall.SyscallN(procGetConsoleScreenBufferInfo.Addr(), uintptr(syscall.Stdout), uintptr(unsafe.Pointer(&info)))
	if errno != 0 {
		err = errno
	}
	width = uint(info.size.x)
	// info.Size.Y includes scrollback
	height = uint(info.window.bottom - info.window.top + 1)
	return
}

// https://learn.microsoft.com/en-us/windows/console/getconsolemode
func state(handle syscall.Handle) (mode uint32, err error) {
	err = syscall.GetConsoleMode(handle, &mode)
	return
}

// https://learn.microsoft.com/en-us/windows/console/setconsolemode
func setState(handle syscall.Handle, mode uint32) error {
	_, _, errno := syscall.SyscallN(procSetConsoleMode.Addr(), uintptr(handle), uintptr(mode))
	if errno != 0 {
		return errno
	}
	return nil
}

func makeRaw() (mode uint32, err error) {
	mode, err = state(syscall.Stdin)
	if err != nil {
		return
	}
	newMode := mode &^ (_ENABLE_ECHO_INPUT | _ENABLE_PROCESSED_INPUT | _ENABLE_LINE_INPUT)
	newMode |= _ENABLE_VIRTUAL_TERMINAL_INPUT
	err = setState(syscall.Stdin, newMode)
	return
}
