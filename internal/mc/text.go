package mc

import (
	"encoding/json"
	"image/color"
	"minefetch/internal/ansi"
	"strconv"
)

// TODO: need a better name; ANSI-like
func FormatLegacy(s string) string {
	var f string

	esc := false
	for _, v := range s {
		if !esc {
			if v == '§' {
				esc = true
			} else {
				f += string(v)
			}
			continue
		} else {
			esc = false
		}

		// TODO: switch to true color via parseColor
		// https://minecraft.wiki/w/Formatting_codes#Java_Edition
		if (v >= '0' && v <= '9') || (v >= 'a' && v <= 'f') {
			f += ansi.Reset + ansi.Color(ParseColor(v))
		} else {
			switch v {
			case 'k':
				f += ansi.Invert
			case 'l':
				f += ansi.Bold
			case 'm':
				f += ansi.Strike
			case 'n':
				f += ansi.Underline
			case 'o':
				f += ansi.Italic
			case 'r':
				f += ansi.Reset
			}
		}
	}

	return f + ansi.Reset
}

// https://minecraft.wiki/w/Text_component_format#Java_Edition
type Text struct {
	Text  string
	Extra []Text
	// NOTE: legacy formatting codes take precedence
	Color         color.NRGBA
	Bold          bool
	Italic        bool
	Underlined    bool
	Strikethrough bool
	Obfuscated    bool
}

func (t Text) Raw() string {
	s := t.Text
	for _, t = range t.Extra {
		s += t.Raw()
	}
	return s
}

func (t Text) Ansi() string {
	s := ansi.Color(t.Color)
	if t.Bold {
		s += ansi.Bold
	}
	if t.Italic {
		s += ansi.Italic
	}
	if t.Underlined {
		s += ansi.Underline
	}
	if t.Strikethrough {
		s += ansi.Strike
	}
	if t.Obfuscated {
		s += ansi.Invert
	}
	s += FormatLegacy(t.Text)
	for _, t = range t.Extra {
		s += t.Ansi()
	}
	return s + ansi.Reset
}

func (t *Text) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*t = NormText(v, Text{})
	return nil
}

func ParseColor(v any) color.NRGBA {
	switch v {
	case '0', "black":
		return color.NRGBA{0, 0, 0, 255}
	case '1', "dark_blue":
		return color.NRGBA{0, 0, 170, 255}
	case '2', "dark_green":
		return color.NRGBA{0, 170, 0, 255}
	case '3', "dark_aqua":
		return color.NRGBA{0, 170, 170, 255}
	case '4', "dark_red":
		return color.NRGBA{170, 0, 0, 255}
	case '5', "dark_purple":
		return color.NRGBA{170, 0, 170, 255}
	case '6', "gold":
		return color.NRGBA{255, 170, 0, 255}
	case '7', "gray":
		return color.NRGBA{170, 170, 170, 255}
	case '8', "dark_gray":
		return color.NRGBA{85, 85, 85, 255}
	case '9', "blue":
		return color.NRGBA{85, 85, 255, 255}
	case 'a', "green":
		return color.NRGBA{85, 255, 85, 255}
	case 'b', "aqua":
		return color.NRGBA{85, 255, 255, 255}
	case 'c', "red":
		return color.NRGBA{255, 85, 85, 255}
	case 'd', "light_purple":
		return color.NRGBA{255, 85, 255, 255}
	case 'e', "yellow":
		return color.NRGBA{255, 255, 85, 255}
	case 'f', "white":
		return color.NRGBA{255, 255, 255, 255}
	}
	if v, ok := v.(string); ok {
		if v[0] == '#' {
			x, err := strconv.ParseUint(v[1:], 16, 32)
			if err == nil {
				return color.NRGBA{uint8(x >> 16), uint8(x >> 8), uint8(x), 255}
			}
		}
	}
	return color.NRGBA{128, 128, 128, 255}
}

func NormText(v any, parent Text) Text {
	if parent.Color == (color.NRGBA{}) {
		parent.Color = ParseColor(nil)
	}
	switch v := v.(type) {
	case string:
		t := parent
		t.Text = v
		t.Extra = []Text{}
		return t
	case []any:
		t := NormText(v[0], parent)
		for _, e := range v[1:] {
			t.Extra = append(t.Extra, NormText(e, t))
		}
		return t
	case map[string]any:
		t := parent
		t.Extra = []Text{}
		if v, ok := v["text"].(string); ok {
			t.Text = v
		} else {
			t.Text = ""
		}
		if v, ok := v["color"].(string); ok {
			t.Color = ParseColor(v)
		}
		if v, ok := v["bold"].(bool); ok {
			t.Bold = v
		}
		if v, ok := v["italic"].(bool); ok {
			t.Italic = v
		}
		if v, ok := v["underlined"].(bool); ok {
			t.Underlined = v
		}
		if v, ok := v["strikethrough"].(bool); ok {
			t.Strikethrough = v
		}
		if v, ok := v["obfuscated"].(bool); ok {
			t.Obfuscated = v
		}
		if v, ok := v["extra"].([]any); ok {
			for _, e := range v {
				t.Extra = append(t.Extra, NormText(e, t))
			}
		}
		return t
	}
	return Text{} // TODO: probably not...
}
