package term

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// RemoveCsi returns s without [Control Sequence Introducer commands].
//
// [Control Sequence Introducer commands]: https://en.wikipedia.org/wiki/ANSI_escape_code#CSIsection
func RemoveCsi(s string) string {
	re := regexp.MustCompile(`\033\[\d+(?:;\d+)*[a-zA-Z]`)
	return re.ReplaceAllString(s, "")
}

// TrimSpace is like strings.TrimSpace, but ignores [Control Sequence Introducer commands].
//
// Any and all commands before the first non-white-space character will be kept.
// No commands following the last non-white-space character are kept.
// Commands that have an effect on white space (e.g. Bg, Strike) are not handled as special cases.
//
// [Control Sequence Introducer commands]: https://en.wikipedia.org/wiki/ANSI_escape_code#CSIsection
func TrimSpace(s string) string {
	var b strings.Builder
	start := -1
	stop := 0
	esc := false
	csi := false
	for i, r := range s {
		switch {
		case csi:
			// break
		case esc:
			esc = false
			csi = r == '['
		case r == 033:
			esc = true
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
	b.WriteString(s[start : stop+i])
	b.WriteString(Reset)

	return b.String()
}
