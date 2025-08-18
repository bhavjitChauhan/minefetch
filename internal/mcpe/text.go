package mcpe

import (
	"image/color"
	"minefetch/internal/term"
	"strconv"
	"strings"
)

// LegacyTextAnsi converts [Minecraft legacy formatting] to ANSI escape codes.
//
// Bedrock Edition does not support the strikethrough and underlined codes.
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
		case 'o':
			b.WriteString(term.Italic)
		case 'r':
			b.WriteString(term.Reset)
		default:
			if (v >= '0' && v <= '9') || (v >= 'a' && v <= 'v') {
				b.WriteString(term.Color(ParseColor(v)))
			}
		}
	}
	b.WriteString(term.Reset)
	return b.String()
}

// Colors corresponding to legacy formatting color codes.
//
// Bedrock Edition supports 11 additional colors compared to the Java Edition.
var (
	Default           = color.NRGBA{128, 128, 128, 255}
	Black             = color.NRGBA{0, 0, 0, 255}
	DarkBlue          = color.NRGBA{0, 0, 170, 255}
	DarkGreen         = color.NRGBA{0, 170, 0, 255}
	DarkAqua          = color.NRGBA{0, 170, 170, 255}
	DarkRed           = color.NRGBA{170, 0, 0, 255}
	DarkPurple        = color.NRGBA{170, 0, 170, 255}
	Gold              = color.NRGBA{255, 170, 0, 255}
	Gray              = color.NRGBA{198, 198, 198, 255}
	DarkGray          = color.NRGBA{85, 85, 85, 255}
	Blue              = color.NRGBA{85, 85, 255, 255}
	Green             = color.NRGBA{85, 255, 85, 255}
	Aqua              = color.NRGBA{85, 255, 255, 255}
	Red               = color.NRGBA{255, 85, 85, 255}
	LightPurple       = color.NRGBA{255, 85, 255, 255}
	Yellow            = color.NRGBA{255, 255, 85, 255}
	White             = color.NRGBA{255, 255, 255, 255}
	MinecoinGold      = color.NRGBA{221, 214, 5, 255}
	MaterialQuartz    = color.NRGBA{227, 212, 209, 255}
	MaterialIron      = color.NRGBA{206, 202, 202, 255}
	MaterialNetherite = color.NRGBA{68, 58, 59, 255}
	MaterialRedstone  = color.NRGBA{151, 22, 7, 255}
	MaterialCopper    = color.NRGBA{180, 104, 77, 255}
	MaterialGold      = color.NRGBA{222, 177, 45, 255}
	MaterialEmerald   = color.NRGBA{17, 159, 54, 255}
	MaterialDiamond   = color.NRGBA{44, 186, 168, 255}
	MaterialLapis     = color.NRGBA{33, 73, 123, 255}
	MaterialAmethyst  = color.NRGBA{154, 92, 198, 255}
	MaterialResin     = color.NRGBA{235, 114, 20, 255}
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
	case 'g', "minecoin_gold":
		return MinecoinGold
	case 'h', "material_quartz":
		return MaterialQuartz
	case 'i', "material_iron":
		return MaterialIron
	case 'j', "material_netherite":
		return MaterialNetherite
	case 'm', "material_redstone":
		return MaterialRedstone
	case 'n', "material_copper":
		return MaterialCopper
	case 'p', "material_gold":
		return MaterialGold
	case 'q', "material_emerald":
		return MaterialEmerald
	case 's', "material_diamond":
		return MaterialDiamond
	case 't', "material_lapis":
		return MaterialLapis
	case 'u', "material_amethyst":
		return MaterialAmethyst
	case 'v', "material_resin":
		return MaterialResin
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
