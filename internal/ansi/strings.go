package ansi

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

func RemoveCsi(s string) string {
	re := regexp.MustCompile(`\033\[\d+(?:;\d+)*[a-zA-Z]`)
	return re.ReplaceAllString(s, "")
}

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
	var b strings.Builder
	for i, r := range s {
		switch {
		case csi:
			break
		case esc:
			esc = false
			csi = r == '['
		case r == 033:
			esc = true
		// TODO: don't trim newlines?
		// TODO: don't trim struck spaces
		case unicode.IsSpace(r):
			continue
		}
		if esc || csi {
			csi = csi && !((r > 'A' && r < 'Z') || (r > 'a' && r < 'z'))
			if start == -1 {
				b.WriteRune(r)
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
	b.WriteString(s[start:stop+i] + Reset)

	return b.String()
}
