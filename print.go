package main

import (
	"bytes"
	"fmt"
	"log"
	"minefetch/internal/ansi"
	"minefetch/internal/image/pngconfig"
	"minefetch/internal/mc"
	"minefetch/internal/mcpe"
	"net"
	"strconv"
	"strings"
	"unicode/utf8"
)

const padding = 2

var lines = 0

func printLine(label string, data any) {
	ss := strings.Split(fmt.Sprint(data), "\n")
	if cfg.icon {
		fmt.Print(ansi.Fwd(cfg.iconSize + padding))
	}
	if len(ss) == 1 && ansi.RemoveCsi(ss[0]) == "" {
		ss[0] = ansi.Gray + "(empty)"
	}
	fmt.Println(ansi.Bold + ansi.Blue + label + ansi.Reset + ": " + ss[0])
	for _, v := range ss[1:] {
		fwd := uint(len(label)) + 2
		if cfg.icon {
			fwd += cfg.iconSize + padding
			fmt.Println(ansi.Fwd(fwd) + v)
		} else {
			fmt.Println(strings.Repeat(" ", int(fwd)) + v)
		}
	}
	fmt.Print(ansi.Reset)
	lines += len(ss)
}

func printErr(label string, err error) {
	printLine(label, ansi.DarkYellow+"Failed "+ansi.Gray+"("+err.Error()+ansi.Gray+")")
}

func printTimeout(label string) {
	printLine(label, ansi.DarkYellow+"Timed out")
}

func printNetInfo() {
	ip := cfg.argHost
	if net.ParseIP(cfg.argHost) == nil {
		ips, err := net.LookupIP(cfg.host)
		if err == nil {
			ip = ips[0].String()
		}
	}
	if cfg.argHost != ip {
		printLine("Host", cfg.argHost)
	}
	if cfg.host != cfg.argHost {
		printLine("SRV", cfg.host)
	}
	if ip != "" {
		printLine("IP", ip)
	}
	port := cfg.port
	if cfg.bedrock {
		port = cfg.bedrockPort
	}
	printLine("Port", port)
	if cfg.crossplay {
		printLine("Bedrock port", cfg.bedrockPort)
	}
}

func printMotd(s string) {
	ss := strings.Split(s, "\n")
	for i, s := range ss {
		ss[i] = ansi.TrimSpace(s)
	}
	if len(ss) > 1 {
		n := [2]int{utf8.RuneCountInString(ansi.RemoveCsi(ss[0])), utf8.RuneCountInString(ansi.RemoveCsi(ss[1]))}
		i := 0
		if n[1] < n[0] {
			i = 1
		}
		j := (i + 1) % 2
		ss[i] = strings.Repeat(" ", (n[j]-n[i])/2) + ss[i]
	}
	printLine("MOTD", strings.Join(ss, "\n"))
}

func printPlayers(online, max int, sample []string) {
	s := fmt.Sprintf("%v"+ansi.Gray+"/"+ansi.Reset+"%v", online, max)
	for _, v := range sample {
		s += "\n" + mc.LegacyTextAnsi(v)
	}
	printLine("Players", s)
}

func printStatus(status *mc.StatusResponse) {
	if cfg.icon {
		printIcon(status.Icon)
	}

	printMotd(status.Motd.Ansi())

	{
		ms := status.Latency.Milliseconds()
		var c string
		switch {
		case ms < 50:
			c = ansi.Green
		case ms < 100:
			c = ansi.Yellow
		default:
			c = ansi.Red
		}
		printLine("Ping", fmt.Sprint(c, ms, " ms"))
	}

	printLine("Version", mc.LegacyTextAnsi(status.Version.Name))

	{
		var sample []string
		for _, p := range status.Players.Sample {
			sample = append(sample, p.Name)
		}
		printPlayers(status.Players.Online, status.Players.Max, sample)
	}

	{
		var s string
		protoVerName, ok := mc.VersionIdName[status.Version.Protocol]
		if ok {
			s = fmt.Sprintf("%v "+ansi.Gray+"(%v)", protoVerName, status.Version.Protocol)
		} else {
			s = strconv.Itoa(int(status.Version.Protocol))
		}
		printLine("Protocol", s)
	}

	if status.Icon != nil {
		iconConfig, _ := pngconfig.DecodeConfig(bytes.NewReader(status.Icon))
		interlaced := ""
		if iconConfig.Interlaced {
			interlaced = "Interlaced "
		}
		printLine("Icon", fmt.Sprintf("%v%v-bit %v", interlaced, iconConfig.BitDepth, formatColorType(iconConfig.ColorType)))
	} else {
		printLine("Icon", "Default")
	}

	printLine("Secure chat", formatBool(!status.EnforcesSecureChat, "Not enforced", "Enforced"))

	if status.PreventsChatReports {
		printLine("Prevents chat reports", ansi.Green+"Yes")

	}
}

func printBedrock(status mcpe.StatusResponse) {
	printLine("Name", mcpe.LegacyTextAnsi(status.Name))
	printLine("Level", mcpe.LegacyTextAnsi(status.Level))
	printLine("Version", fmt.Sprintf("%v "+ansi.Gray+"(%v)", status.Version.Name, status.Version.Protocol))
	printPlayers(status.Players.Online, status.Players.Max, nil)
	printLine("Edition", status.Edition)
	printLine("Game Mode", fmt.Sprintf("%v "+ansi.Gray+"(%v)", status.GameMode.Name, status.GameMode.ID))
}

func printQuery(query mc.QueryResponse) {
	prev := lines
	if !cfg.status {
		printMotd(mc.LegacyTextAnsi(query.Motd))
		// TODO: ping
		printLine("Version", mc.LegacyTextAnsi(query.Version))
		printPlayers(query.Players.Online, query.Players.Max, query.Players.Sample)
	}
	if query.Software != "" {
		printLine("Software", query.Software)
	}
	if len(query.Plugins) > 0 {
		printLine("Plugins", strings.Join(query.Plugins, "\n"))
	}
	if lines == prev {
		printLine("Query", ansi.Green+"Enabled")
	}
}

func printResult[T any](result result, label string, fn func(T), failed string) {
	switch {
	case result.v != nil:
		v, ok := result.v.(T)
		if !ok {
			log.Panicf("Unexpected result value for %v: %v", label, result.v)
		}
		fn(v)
	case result.err != nil, result.timeout:
		if failed != "" {
			printLine(label, failed)
		} else if result.timeout {
			printTimeout(label)
		} else {
			printErr(label, result.err)
		}
	default:
		log.Panicln("Unexpected result state:", result)
	}
}

func printResults(results results) {
	if cfg.icon && (!cfg.status || results[resultStatus].err != nil || results[resultStatus].timeout) {
		printIcon(nil)
	}

	if cfg.status {
		result := results[resultStatus]
		if result.err != nil || result.timeout {
			cfg.status = false
		}
		printResult(results[resultStatus], "Java", func(status mc.StatusResponse) {
			printStatus(&status)
		}, ansi.Red+"Offline")
	}

	if cfg.crossplay && results[resultStatus].v == nil && results[resultBedrockStatus].v != nil {
		cfg.bedrock = true
		cfg.crossplay = false
	}
	if cfg.bedrock {
		printResult(results[resultBedrockStatus], "Bedrock", func(status mcpe.StatusResponse) {
			printBedrock(status)
		}, ansi.Red+"Offline")
	}

	if cfg.query {
		result := results[resultQuery]
		printResult(result, "Query", func(query mc.QueryResponse) {
			printQuery(query)
		}, ansi.Red+"Disabled")
	}

	if cfg.crossplay {
		printResult(results[resultBedrockStatus], "Crossplay", func(status mcpe.StatusResponse) {
			printLine("Crossplay", ansi.Green+"Yes")
		}, ansi.Red+"No")
		if results[resultBedrockStatus].v == nil {
			cfg.crossplay = false
		}
	}

	printNetInfo()

	if cfg.blocked {
		printResult(results[resultBlocked], "Blocked", func(blocked string) {
			printLine("Blocked", formatBool(blocked == "", "No", fmt.Sprintf("Yes %v(%v)", ansi.Gray, blocked)))
		}, "")
	}

	if cfg.cracked {
		printResult(results[resultCracked], "Cracked", func(crackedWhitelisted [2]bool) {
			printLine("Cracked", formatBool(crackedWhitelisted[0], ansi.Reset+"Yes", ansi.Reset+"No"))
			if crackedWhitelisted[0] {
				printLine("Whitelist", formatBool(!crackedWhitelisted[1], "Off", "On"))
			}
		}, "")
	}

	if cfg.rcon {
		printResult(results[resultRcon], "RCON", func(enabled bool) {
			printLine("RCON", formatBool(!enabled, "Disabled", "Enabled"))
		}, "")
	}

	if cfg.palette {
		printPalette()
	}

	if cfg.icon && lines < int(iconHeight())+1 {
		fmt.Print(strings.Repeat("\n", int(iconHeight())-lines+1))
	} else {
		fmt.Print("\n")
	}
}

func printPalette() {
	const codes = "0123456789abcdef"
	var b strings.Builder
	b.WriteRune('\n')
	if cfg.icon {
		b.WriteString(ansi.Fwd(cfg.iconSize + padding))
	}
	for _, code := range codes[:len(codes)/2] {
		b.WriteString(ansi.Bg(mc.ParseColor(code)))
		b.WriteString("   ")
	}
	b.WriteString(ansi.Reset)
	b.WriteRune('\n')
	if cfg.icon {
		b.WriteString(ansi.Fwd(cfg.iconSize + padding))
	}
	for _, code := range codes[len(codes)/2:] {
		b.WriteString(ansi.Bg(mc.ParseColor(code)))
		b.WriteString("   ")
	}
	lines += 3
	if cfg.bedrock {
		const codes = "ghijmnpqstuv"
		b.WriteString(ansi.Reset)
		b.WriteRune('\n')
		if cfg.icon {
			b.WriteString(ansi.Fwd(cfg.iconSize + padding))
		}
		for _, code := range codes {
			b.WriteString(ansi.Bg(mcpe.ParseColor(code)))
			b.WriteString("  ")
		}
		lines++
	}
	b.WriteString(ansi.Reset)
	b.WriteRune('\n')
	fmt.Print(b.String())
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
