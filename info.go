package main

import (
	"encoding/base64"
	"fmt"
	"minefetch/internal/ansi"
	"minefetch/internal/image/pngconfig"
	"minefetch/internal/mc"
	"net"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"
)

const padding = 2

var lines = 0

type info struct {
	label string
	data  any
}

func printInfo(i info) {
	s := strings.Split(fmt.Sprint(i.data), "\n")
	fmt.Println(ansi.Fwd(iconWidth+padding) + ansi.Bold + ansi.Blue + i.label + ansi.Reset + ": " + s[0])
	for _, v := range s[1:] {
		fmt.Println(ansi.Fwd(iconWidth+padding+uint(len(i.label))+2) + v)
	}
	fmt.Print(ansi.Reset)
	lines += len(s)
}

func printStatus(host string, port uint16, status *mc.StatusResponse) {
	var ii []info

	{
		ss := strings.Split(status.Description.Ansi(), "\n")
		for i, s := range ss {
			ss[i] = ansi.TrimSpace(s)
		}
		if len(ss) > 1 {
			runeCounts := [2]int{utf8.RuneCountInString(ansi.RemoveCsi(ss[0])), utf8.RuneCountInString(ansi.RemoveCsi(ss[1]))}
			i := 0
			if runeCounts[1] < runeCounts[0] {
				i = 1
			}
			j := (i + 1) % 2
			ss[i] = strings.Repeat(" ", (runeCounts[j]-runeCounts[i])/2) + ss[i]
		}
		ii = append(ii, info{"MOTD", strings.Join(ss, "\n")})
	}

	ii = append(ii, info{"Ping", fmt.Sprint(status.Latency.Milliseconds(), " ms")})

	ii = append(ii, info{"Version", mc.LegacyTextAnsi(status.Version.Name)})

	{
		s := fmt.Sprintf("%v"+ansi.Gray+"/"+ansi.Reset+"%v", status.Players.Online, status.Players.Max)
		for _, v := range status.Players.Sample {
			s += "\n" + mc.LegacyTextAnsi(v.Name)
		}
		ii = append(ii, info{"Players", s})
	}

	{
		argHost, _, err := net.SplitHostPort(os.Args[1])
		if err != nil {
			argHost = os.Args[1]
		}

		ip := argHost
		if net.ParseIP(argHost) == nil {
			ips, err := net.LookupIP(host)
			if err == nil {
				ip = ips[0].String()
			}
		}
		if argHost != ip {
			ii = append(ii, info{"Host", argHost})
		}
		if host != argHost {
			ii = append(ii, info{"SRV", host})
		}
		if ip != "" {
			ii = append(ii, info{"IP", ip})
		}
	}

	ii = append(ii, info{"Port", port})

	{
		var s string
		protoVerName, ok := mc.ProtoVerName[status.Version.Protocol]
		if ok {
			s = fmt.Sprintf("%v "+ansi.Gray+"(%v)", protoVerName, status.Version.Protocol)
		} else {
			s = strconv.Itoa(int(status.Version.Protocol))
		}

		ii = append(ii, info{"Protocol", s})
	}

	if status.Favicon.Image != nil {
		iconConfig, _ := pngconfig.DecodeConfig(base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.TrimPrefix(status.Favicon.Raw, "data:image/png;base64,"))))
		interlaced := ""
		if iconConfig.Interlaced {
			interlaced = "Interlaced "
		}
		ii = append(ii, info{"Icon", fmt.Sprintf("%v%v-bit %v", interlaced, iconConfig.BitDepth, formatColorType(iconConfig.ColorType))})
	} else {
		ii = append(ii, info{"Icon", "Default"})
	}

	ii = append(ii, info{"Secure chat", formatBool(!status.EnforcesSecureChat, "Not enforced", "Enforced")})

	if status.PreventsChatReports {
		ii = append(ii, info{"Prevents chat reports", ansi.Green + "Yes"})

	}

	for _, i := range ii {
		printInfo(i)
	}
}

func printQuery(query *mc.QueryResponse) {
	var ii []info

	ii = append(ii, info{"Query", formatBool(query != nil, "Enabled", "Disabled")})
	if query != nil {
		if query.Software != "" {
			ii = append(ii, info{"Software", query.Software})
		}

		if len(query.Plugins) > 0 {
			ii = append(ii, info{"Plugins", strings.Join(query.Plugins, "\n")})
		}
	}

	for _, i := range ii {
		printInfo(i)
	}
}

func printPalette() {
	const codes = "0123456789abcdef"
	fmt.Print("\n" + ansi.Fwd(iconWidth+2))
	for i, code := range codes {
		fmt.Print(ansi.Bg(mc.ParseColor(code)) + "   ")
		if (i + 1) == (len(codes) / 2) {
			fmt.Print(ansi.Reset + "\n" + ansi.Fwd(iconWidth+2))
		}
	}
	fmt.Println(ansi.Reset)
	lines += 3
}

func formatBool(bool bool, t, f string) string {
	var s string
	if bool {
		s = ansi.Green + t
	} else {
		s = ansi.Red + f
	}
	return s
}

func formatColorType(t pngconfig.ColorType) string {
	switch t {
	case pngconfig.ColorTypeGray:
		return "grayscale"
	case pngconfig.ColorTypeRGB:
		return "RGB"
	case pngconfig.ColorTypeIndexed:
		return "indexed"
	case pngconfig.ColorTypeGrayA:
		return "grayscale + alpha"
	case pngconfig.ColorTypeRGBA:
		return "RGBA"
	}
	return "unknown"
}
