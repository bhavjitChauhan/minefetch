//go:build linux || darwin

package term

import (
	"syscall"
	"unsafe"
)

func size() (width, height uint, err error) {
	// https://man7.org/linux/man-pages/man2/TIOCSWINSZ.2const.html
	type winsize struct {
		row    uint16
		col    uint16
		xpixel uint16
		ypixel uint16
	}

	var w winsize
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdout), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&w)))
	if errno != 0 {
		err = errno
	}
	width = uint(w.col)
	height = uint(w.row)
	return
}

func state() (termios termios, err error) {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdout), ioctlReadTermios, uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		err = errno
	}
	return
}

func setState(termios termios) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(syscall.Stdout), ioctlWriteTermios, uintptr(unsafe.Pointer(&termios)))
	if errno != 0 {
		return errno
	}
	return nil
}

func makeRaw() (termios termios, err error) {
	termios, err = state()
	if err != nil {
		return
	}
	newTermios := termios
	newTermios.iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	newTermios.oflag &^= syscall.OPOST
	newTermios.lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	newTermios.cflag &^= syscall.CSIZE | syscall.PARENB
	newTermios.cflag |= syscall.CS8
	newTermios.cc[syscall.VMIN] = 1
	newTermios.cc[syscall.VTIME] = 0
	err = setState(newTermios)
	return
}
