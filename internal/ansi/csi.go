/*
Package ansi provides utilities to control various aspects of terminals using [ANSI escape codes].

Virtually ever modern terminal emulator supports these sequences.

[ANSI escape sequences]: https://en.wikipedia.org/wiki/ANSI_escape_code
*/
package ansi

import "strconv"

// Up moves the cursor up n cells.
func Up(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "A"
}

// Down moves the cursor down n cells.
func Down(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "B"
}

// Fwd moves the cursor forward n cells.
func Fwd(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "C"
}

// Back moves the cursor back n cells.
func Back(n uint) string {
	return "\033[" + strconv.Itoa(int(n)) + "D"
}
