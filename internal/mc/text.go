package mc

import (
	"encoding/json"
	"image/color"
	"minefetch/internal/ansi"
	"minefetch/internal/emoji"
	"strconv"
)

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
	s += LegacyTextAnsi(emoji.ReplaceColored(t.Text))
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
	*t = normText(v, Text{})
	return nil
}

func normText(v any, parent Text) Text {
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
		t := normText(v[0], parent)
		for _, e := range v[1:] {
			t.Extra = append(t.Extra, normText(e, t))
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
				t.Extra = append(t.Extra, normText(e, t))
			}
		}
		return t
	}
	return Text{} // TODO: probably not...
}

func LegacyTextAnsi(s string) string {
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

var (
	Default     = color.NRGBA{128, 128, 128, 255}
	Black       = color.NRGBA{0, 0, 0, 255}
	DarkBlue    = color.NRGBA{0, 0, 170, 255}
	DarkGreen   = color.NRGBA{0, 170, 0, 255}
	DarkAqua    = color.NRGBA{0, 170, 170, 255}
	DarkRed     = color.NRGBA{170, 0, 0, 255}
	DarkPurple  = color.NRGBA{170, 0, 170, 255}
	Gold        = color.NRGBA{255, 170, 0, 255}
	Gray        = color.NRGBA{170, 170, 170, 255}
	DarkGray    = color.NRGBA{85, 85, 85, 255}
	Blue        = color.NRGBA{85, 85, 255, 255}
	Green       = color.NRGBA{85, 255, 85, 255}
	Aqua        = color.NRGBA{85, 255, 255, 255}
	Red         = color.NRGBA{255, 85, 85, 255}
	LightPurple = color.NRGBA{255, 85, 255, 255}
	Yellow      = color.NRGBA{255, 255, 85, 255}
	White       = color.NRGBA{255, 255, 255, 255}
)

func ParseColor(v any) color.NRGBA {
	switch v {
	case '0', "black":
		return Black
	case '1', "dark_blue":
		return DarkBlue
	case '2', "dark_green":
		return DarkGreen
	case '3', "dark_aqua":
		return DarkAqua
	case '4', "dark_red":
		return DarkRed
	case '5', "dark_purple":
		return DarkPurple
	case '6', "gold":
		return Gold
	case '7', "gray":
		return Gray
	case '8', "dark_gray":
		return DarkGray
	case '9', "blue":
		return Blue
	case 'a', "green":
		return Green
	case 'b', "aqua":
		return Aqua
	case 'c', "red":
		return Red
	case 'd', "light_purple":
		return LightPurple
	case 'e', "yellow":
		return Yellow
	case 'f', "white":
		return White
	}
	if v, ok := v.(string); ok {
		if v != "" && v[0] == '#' {
			x, err := strconv.ParseUint(v[1:], 16, 32)
			if err == nil {
				return color.NRGBA{uint8(x >> 16), uint8(x >> 8), uint8(x), 255}
			}
		}
	}
	return Default
}
