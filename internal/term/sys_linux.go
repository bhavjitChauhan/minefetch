//go:build linux

package term

import (
	"syscall"
)

type termios struct {
	iflag  uint32
	oflag  uint32
	cflag  uint32
	lflag  uint32
	line   uint8
	cc     [19]uint8
	ispeed uint32
	ospeed uint32
}

const ioctlReadTermios = syscall.TCGETS
const ioctlWriteTermios = syscall.TCSETS
