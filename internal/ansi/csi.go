/*
Package ansi provides utilities to control various aspects of terminals using [ANSI escape codes].

Virtually every modern terminal emulator supports these sequences.

[ANSI escape sequences]: https://en.wikipedia.org/wiki/ANSI_escape_code
*/
package ansi

import "strconv"

// Up moves the cursor up n cells.
//
// Returns an empty string if no color support is set.
func Up(n uint) string {
	if ColorSupport == NoColorSupport {
		return ""
	}
	return "\033[" + strconv.Itoa(int(n)) + "A"
}

// Down moves the cursor down n cells.
func Down(n uint) string {
	if ColorSupport == NoColorSupport {
		return ""
	}
	return "\033[" + strconv.Itoa(int(n)) + "B"
}

// Fwd moves the cursor forward n cells.
func Fwd(n uint) string {
	if ColorSupport == NoColorSupport {
		return ""
	}
	return "\033[" + strconv.Itoa(int(n)) + "C"
}

// Back moves the cursor back n cells.
func Back(n uint) string {
	if ColorSupport == NoColorSupport {
		return ""
	}
	return "\033[" + strconv.Itoa(int(n)) + "D"
}
