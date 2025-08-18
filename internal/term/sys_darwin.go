package term

import "syscall"

type termios struct {
	iflag  uint64
	oflag  uint64
	cflag  uint64
	lflag  uint64
	cc     [20]uint8
	ispeed uint64
	ospeed uint64
}

const ioctlReadTermios = syscall.TIOCGETA
const ioctlWriteTermios = syscall.TIOCSETA
