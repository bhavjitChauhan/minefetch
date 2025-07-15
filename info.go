package main

import (
	"encoding/base64"
	"fmt"
	"minefetch/internal/ansi"
	"minefetch/internal/image/pngconfig"
	"minefetch/internal/mc"
	"net"
	"os"
	"strings"
	"time"
)

type infoEntry struct {
	label string
	v     any
}

func printInfo(host string, port uint16, conn net.Conn, latency time.Duration, status *mc.Status) {
	var entries []infoEntry

	desc := strings.Split(status.Description.Ansi(), "\n")
	for i, s := range desc {
		desc[i] = ansi.TrimSpace(s)
	}

	players := fmt.Sprintf("%v"+ansi.Gray+"/"+ansi.Reset+"%v", status.Players.Online, status.Players.Max)
	for _, v := range status.Players.Sample {
		players += "\n" + mc.LegacyTextAnsi(v.Name)
	}

	argHost, _, err := net.SplitHostPort(os.Args[1])
	if err != nil {
		argHost = os.Args[1]
	}

	ip := argHost
	if net.ParseIP(argHost) == nil {
		ip, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	}

	protoVerName := mc.ProtoVerName[status.Version.Protocol]
	if protoVerName != "" {
		protoVerName = " " + ansi.Gray + "(" + protoVerName + ")"
	}

	iconConfig, _ := pngconfig.DecodeConfig(base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.TrimPrefix(status.Favicon.Raw, "data:image/png;base64,"))))

	entries = append(entries, infoEntry{"MOTD", strings.Join(desc, "\n")})
	entries = append(entries, infoEntry{"Ping", fmt.Sprint(latency.Milliseconds(), " ms")})
	entries = append(entries, infoEntry{"Version", mc.LegacyTextAnsi(status.Version.Name)})
	entries = append(entries, infoEntry{"Players", players})
	if argHost != ip {
		entries = append(entries, infoEntry{"Host", argHost})
	}
	if host != argHost {
		entries = append(entries, infoEntry{"SRV", host})
	}
	if ip != "" {
		entries = append(entries, infoEntry{"IP", ip})
	}
	entries = append(entries, infoEntry{"Port", port})
	entries = append(entries, infoEntry{"Protocol", fmt.Sprint(status.Version.Protocol, protoVerName)})
	if status.Favicon.Image != nil {
		interlaced := ""
		if iconConfig.Interlaced {
			interlaced = "Interlaced "
		}
		entries = append(entries, infoEntry{"Icon", fmt.Sprintf("%v%v-bit %v", interlaced, iconConfig.BitDepth, colorTypeString(iconConfig.ColorType))})
	} else {
		entries = append(entries, infoEntry{"Icon", "Default"})
	}
	if status.PreventsChatReports {
		entries = append(entries, infoEntry{"Prevents chat reports", status.PreventsChatReports})
	}

	fmt.Print(ansi.Up(iconHeight-1), ansi.Back(iconWidth))

	for _, e := range entries {
		s := strings.Split(fmt.Sprint(e.v), "\n")
		fmt.Println(ansi.Fwd(iconWidth+2) + ansi.Bold + ansi.Blue + e.label + ansi.Reset + ": " + s[0])
		for _, v := range s[1:] {
			fmt.Println(ansi.Fwd(iconWidth+2+uint(len(e.label))+2) + v)
		}
		fmt.Print(ansi.Reset)
	}

	nl := 0
	for _, e := range entries {
		if s, ok := e.v.(string); ok {
			nl += strings.Count(s, "\n")
		}
	}
	if len(entries)+nl < iconHeight+1 {
		fmt.Print(strings.Repeat("\n", iconHeight-len(entries)-nl+1))
	}
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
