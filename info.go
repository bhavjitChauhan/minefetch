package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"minefetch/internal/ansi"
	"minefetch/internal/image/pngconfig"
	"minefetch/internal/mc"
	"net"
	"os"
	"strconv"
	"strings"
)

type info struct {
	label string
	v     any
}

func printStatus(host string, port uint16, status *mc.StatusResponse, query *mc.QueryResponse) {
	var entries []info

	{
		ss := strings.Split(status.Description.Ansi(), "\n")
		for i, s := range ss {
			ss[i] = ansi.TrimSpace(s)
		}
		entries = append(entries, info{"MOTD", strings.Join(ss, "\n")})
	}

	entries = append(entries, info{"Ping", fmt.Sprint(status.Latency.Milliseconds(), " ms")})

	entries = append(entries, info{"Version", mc.LegacyTextAnsi(status.Version.Name)})

	{
		s := fmt.Sprintf("%v"+ansi.Gray+"/"+ansi.Reset+"%v", status.Players.Online, status.Players.Max)
		for _, v := range status.Players.Sample {
			s += "\n" + mc.LegacyTextAnsi(v.Name)
		}
		entries = append(entries, info{"Players", s})
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
			entries = append(entries, info{"Host", argHost})
		}
		if host != argHost {
			entries = append(entries, info{"SRV", host})
		}
		if ip != "" {
			entries = append(entries, info{"IP", ip})
		}
	}

	entries = append(entries, info{"Port", port})

	{
		var s string
		protoVerName, ok := mc.ProtoVerName[status.Version.Protocol]
		if ok {
			s = fmt.Sprintf("%v "+ansi.Gray+"(%v)", protoVerName, status.Version.Protocol)
		} else {
			s = strconv.Itoa(int(status.Version.Protocol))
		}

		entries = append(entries, info{"Protocol", s})
	}

	if status.Favicon.Image != nil {
		iconConfig, _ := pngconfig.DecodeConfig(base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.TrimPrefix(status.Favicon.Raw, "data:image/png;base64,"))))
		interlaced := ""
		if iconConfig.Interlaced {
			interlaced = "Interlaced "
		}
		entries = append(entries, info{"Icon", fmt.Sprintf("%v%v-bit %v", interlaced, iconConfig.BitDepth, colorTypeString(iconConfig.ColorType))})
	} else {
		entries = append(entries, info{"Icon", "Default"})
	}

	{
		s := "Not enforced"
		if status.EnforcesSecureChat {
			s = "Enforced"
		}
		entries = append(entries, info{"Secure chat", s})
	}

	if status.PreventsChatReports {
		entries = append(entries, info{"Prevents chat reports", status.PreventsChatReports})

	}

	if query != nil {
		entries = append(entries, info{"Query", "Enabled"})

		if query.Software != "" {
			entries = append(entries, info{"Software", query.Software})
		}

		if len(query.Plugins) > 0 {
			entries = append(entries, info{"Plugins", strings.Join(query.Plugins, "\n")})
		}
	} else {
		entries = append(entries, info{"Query", "Disabled"})
	}

	{
		err := printIcon(&status.Favicon)
		if err != nil {
			log.Fatalln("Failed to print icon:", err)
		}
	}

	fmt.Print(ansi.Up(iconHeight-1) + ansi.Back(iconWidth))

	for _, e := range entries {
		printEntry(e)
	}

	lines := countLines(entries)
	if lines < iconHeight+1 {
		fmt.Print(strings.Repeat("\n", iconHeight-lines+1))
	}
}

func printEntry(e info) {
	s := strings.Split(fmt.Sprint(e.v), "\n")
	fmt.Println(ansi.Fwd(iconWidth+2) + ansi.Bold + ansi.Blue + e.label + ansi.Reset + ": " + s[0])
	for _, v := range s[1:] {
		fmt.Println(ansi.Fwd(iconWidth+2+uint(len(e.label))+2) + v)
	}
	fmt.Print(ansi.Reset)
}

func countLines(entries []info) (lines int) {
	lines = len(entries)
	for _, e := range entries {
		if s, ok := e.v.(string); ok {
			lines += strings.Count(s, "\n")
		}
	}
	return
}

func colorTypeString(t pngconfig.ColorType) string {
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
