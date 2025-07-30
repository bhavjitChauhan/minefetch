/*
Package ansi provides utilities to control various aspects of terminals using [ANSI escape codes].

Virtually ever modern terminal emulator supports these sequences.

[ANSI escape sequences]: https://en.wikipedia.org/wiki/ANSI_escape_code
*/
package ansi

import "strconv"

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
