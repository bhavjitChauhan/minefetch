/*
Package flag is like [spf13/pflag], but doesn't follow Go's flag and doesn't
strictly adhere to any standard.

There are no "boolean" flags, only toggles. If a toggle flag is present, the
boolean value will be set to the opposite of the default, if present, or true.

*Flag runes* are single-character flags.

*Flag strings* consist of multiple flag runes. Flag strings may only contain
toggle flags.

[sp13/pflag]: https://github.com/spf13/pflag
*/
package flag

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Flag struct {
	Name     string
	Rune     rune
	Usage    string
	Value    any
	DefValue any
}

var Flags []Flag

func Var(p any, name string, r rune, v any, usage string) {
	Flags = append(Flags, Flag{
		name,
		r,
		usage,
		p,
		v,
	})
}

func Parse() (remaining []string, err error) {
	flagNameMap := make(map[string]Flag)
	flagRuneMap := make(map[rune]Flag)
	for _, f := range Flags {
		if f.Name != "" {
			flagNameMap[f.Name] = f
		}
		if f.Rune != 0 {
			flagRuneMap[f.Rune] = f
		}
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		skip := false
		var v string
		if j := strings.IndexRune(arg, '='); j != -1 {
			v = arg[j+1:]
			arg = arg[:j]
		} else if i+1 < len(args) && !isFlag(args[i+1]) {
			v = args[i+1]
			skip = true
		}

		var f Flag
		switch {
		case isFlagName(arg):
			name := arg[2:]
			var ok bool
			f, ok = flagNameMap[name]
			if !ok {
				err = errors.New("unknown flag name: " + name)
				return
			}
		case isFlagString(arg):
			for _, r := range arg[1:] {
				f, ok := flagRuneMap[r]
				if !ok {
					err = errors.New("unknown boolean flag rune: " + string(r))
					return
				}
				p, ok := f.Value.(*bool)
				if !ok {
					err = fmt.Errorf("unexpected non-boolean flag type: %T", f.Value)
					return
				}
				v, _ := f.DefValue.(bool)
				*p = !v
			}
			continue
		case isFlagRune(arg):
			r := rune(arg[1])
			var ok bool
			f, ok = flagRuneMap[r]
			if !ok {
				err = errors.New("unknown flag rune: " + string(r))
				return
			}
		default:
			remaining = append(remaining, arg)
			continue
		}

		switch p := f.Value.(type) {
		case *bool:
			v, _ := f.DefValue.(bool)
			*p = !v
			continue
		case *uint:
			var uint64 uint64
			uint64, err = strconv.ParseUint(v, 10, 0)
			if err != nil {
				err = fmt.Errorf("failed to parse integer flag value: %v", v)
				return
			}
			*p = uint(uint64)
		case *time.Duration:
			var d time.Duration
			d, err = time.ParseDuration(v)
			if err != nil {
				err = fmt.Errorf("failed to parse duration flag value: %v", v)
				return
			}
			*p = d
		case *string:
			*p = v
		default:
			return nil, fmt.Errorf("cannot parse flag type: %T", f.Value)
		}

		if skip {
			i++
		}
	}

	return
}

// Inspired by [spf13/pflag].
//
// [pflag]: https://github.com/spf13/pflag/blob/c78f730fb023e4012c4097b24408867cd5c5bdde/flag.go#L697
func Print() {
	lines := make([]string, 0, len(Flags))
	max := 0
	for _, f := range Flags {
		b := strings.Builder{}
		if f.Rune != 0 {
			b.WriteString(fmt.Sprintf("    -%s, ", string(f.Rune)))
		} else {
			b.WriteString("        ")
		}
		isBool := false
		var s string
		switch f.DefValue.(type) {
		case bool:
			s = f.Name
			isBool = true
		case time.Duration:
			s = f.Name + " duration"
		default:
			s = fmt.Sprintf("%s %T", f.Name, f.DefValue)
		}
		b.WriteString(fmt.Sprintf("--%s\x00%s", s, f.Usage))
		if !isBool {
			b.WriteString(fmt.Sprintf(" (default %v)", f.DefValue))
		}
		lines = append(lines, b.String())
		if len := len(s); len > max {
			max = len
		}
	}
	b := strings.Builder{}
	for _, line := range lines {
		i := strings.IndexRune(line, 0)
		b.WriteString(line[:i])
		b.WriteString(strings.Repeat(" ", max-i+12)) // indent (10) + padding (2)
		b.WriteString(line[i+1:])
		b.WriteRune('\n')
	}
	fmt.Print(b.String())
}

func isFlagName(s string) bool {
	if len(s) < 3 {
		return false
	}
	return s[0] == '-' && s[1] == '-' && unicode.IsLetter(rune(s[2]))
}

func isFlagRune(s string) bool {
	if len(s) != 2 {
		return false
	}
	return s[0] == '-' && unicode.IsLetter(rune(s[1]))
}

func isFlagString(s string) bool {
	if len(s) < 3 {
		return false
	}
	return s[0] == '-' && unicode.IsLetter(rune(s[1])) && unicode.IsLetter(rune(s[2]))
}

func isFlag(s string) bool {
	if len(s) < 2 {
		return false
	}
	return s[0] == '-' && (s[1] == '-' || unicode.IsLetter(rune(s[1])))
}
