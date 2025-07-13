package ansi

import (
	"unicode"
	"unicode/utf8"
)

// Like strings.TrimSpace, but ignores [Control Sequence Introducer commands].
//
// Trailing commands are also trimmed, so this function may result in undesired
// behavior for strings utilizing, e.g., the background color.
//
// [Control Sequence Introducer commands]: https://en.wikipedia.org/wiki/ANSI_escape_code#CSIsection
func TrimSpace(s string) string {
	start := -1
	stop := 0
	esc := false
	csi := false
	ss := ""
	for i, r := range s {
		switch {
		case csi:
			csi = !((r > 'A' && r < 'Z') || (r > 'a' && r < 'z'))
			if start == -1 {
				ss += string(r)
			}
			continue
		case esc:
			esc = false
			csi = r == '['
		case r == 033:
			esc = true
		// TODO: don't trim newlines?
		case unicode.IsSpace(r):
			continue
		}
		if esc || csi {
			if start == -1 {
				ss += string(r)
			}
			continue
		}
		if start == -1 {
			start = i
		}
		stop = i
	}

	// All-space string
	if start == -1 {
		return s
	}

	// Handle multi-byte UTF-8 characters
	i := 4
	if stop+i > len(s) {
		i = len(s) - stop
	}
	for utf8.RuneCountInString(s[stop:stop+i]) > 1 {
		i--
	}
	ss += s[start:stop+i] + Reset

	return ss
}
