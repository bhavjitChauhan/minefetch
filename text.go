package main

func formatLegacy(format string) (string, error) {
	var s string

	esc := false
	for _, v := range format {
		if !esc {
			if v == '§' {
				esc = true
			} else {
				s += string(v)
			}
			continue
		} else {
			esc = false
		}

		// https://minecraft.wiki/w/Formatting_codes#Java_Edition
		if (v >= '0' && v <= '9') || (v >= 'a' && v <= 'f') {
			s += reset
		}

		switch v {
		case '0':
			s += black
		case '1':
			s += blue
		case '2':
			s += green
		case '3':
			s += cyan
		case '4':
			s += red
		case '5':
			s += magenta
		case '6':
			s += yellow
		case '7':
			s += white
		case '8':
			s += brightBlack
		case '9':
			s += brightBlue
		case 'a':
			s += brightGreen
		case 'b':
			s += brightCyan
		case 'c':
			s += brightRed
		case 'd':
			s += brightMagenta
		case 'e':
			s += brightYellow
		case 'f':
			s += brightWhite

		case 'k':
			s += invert
		case 'l':
			s += bold
		case 'm':
			s += strike
		case 'n':
			s += underline
		case 'o':
			s += italic
		case 'r':
			s += reset
		}
	}
	s += reset

	return s, nil
}
