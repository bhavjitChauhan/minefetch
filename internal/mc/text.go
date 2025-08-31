package mc

import (
	"encoding/json"
	"image/color"
	"strconv"
	"strings"

	"github.com/bhavjitChauhan/minefetch/internal/emoji"
	"github.com/bhavjitChauhan/minefetch/internal/term"
)

// Text is a text component format object.
//
// https://minecraft.wiki/w/Text_component_format#Java_Edition
type Text struct {
	Text          string
	Extra         []Text
	Color         color.NRGBA
	Bold          bool
	Italic        bool
	Underlined    bool
	Strikethrough bool
	Obfuscated    bool
}

// Raw returns all of t's descendents' Text fields flattened to a string.
func (t Text) Raw() string {
	var b strings.Builder
	b.WriteString(t.Text)
	for _, t = range t.Extra {
		b.WriteString(t.Raw())
	}
	return b.String()
}

// Ansi returns a representation of t using ANSI escape codes.
func (t Text) Ansi() string {
	var b strings.Builder
	b.WriteString(term.Color(t.Color))
	if t.Bold {
		b.WriteString(term.Bold)
	}
	if t.Italic {
		b.WriteString(term.Italic)
	}
	if t.Underlined {
		b.WriteString(term.Underline)
	}
	if t.Strikethrough {
		b.WriteString(term.Strike)
	}
	if t.Obfuscated {
		b.WriteString(term.Invert)
	}
	b.WriteString(LegacyTextAnsi(emoji.ReplaceColored(t.Text)))
	for _, t = range t.Extra {
		b.WriteString(t.Ansi())
	}
	b.WriteString(term.Reset)
	return b.String()
}

func (t *Text) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	*t = normText(v, Text{})
	return nil
}

// normText normalizes v to a Text struct.
// Inheritance is precomputed; descendant components have their final text component formatting.
//
// v may be a string, text component object or a list of either.
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

// LegacyTextAnsi converts [Minecraft legacy formatting] to ANSI escape codes.
//
// [Minecraft legacy formatting]: https://minecraft.wiki/w/Formatting_codes
func LegacyTextAnsi(s string) string {
	var b strings.Builder
	esc := false
	for _, v := range s {
		if !esc {
			if v == '§' {
				esc = true
			} else {
				b.WriteRune(v)
			}
			continue
		} else {
			esc = false
		}

		switch v {
		case 'k':
			b.WriteString(term.Invert)
		case 'l':
			b.WriteString(term.Bold)
		case 'm':
			b.WriteString(term.Strike)
		case 'n':
			b.WriteString(term.Underline)
		case 'o':
			b.WriteString(term.Italic)
		case 'r':
			b.WriteString(term.Reset)
		default:
			if (v >= '0' && v <= '9') || (v >= 'a' && v <= 'f') {
				b.WriteString(term.Reset)
				b.WriteString(term.Color(ParseColor(v)))
			}
		}
	}
	b.WriteString(term.Reset)
	return b.String()
}

// Colors corresponding to legacy formatting color codes and the server list default text color.
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

// ParseColor converts v to a Minecraft color.
//
// v may be a legacy formatting code, named color or hex color code.
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
