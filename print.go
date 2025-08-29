package main

import (
	"bytes"
	"fmt"
	"minefetch/internal/image/pngconfig"
	"minefetch/internal/mc"
	"minefetch/internal/mcpe"
	"minefetch/internal/term"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const padding = 2

var lines = 0

func printLine(label string, data any) {
	ss := strings.Split(fmt.Sprint(data), "\n")
	if cfg.icon.enabled {
		fmt.Print(term.Fwd(cfg.icon.size + padding))
	}
	if len(ss) == 1 && term.RemoveCsi(ss[0]) == "" {
		ss[0] = term.Gray + "(empty)"
	}
	fmt.Println(term.Bold + term.Blue + label + term.Reset + ": " + ss[0] + term.Reset)
	for _, v := range ss[1:] {
		fwd := uint(len(label)) + 2
		if cfg.icon.enabled && term.ColorSupport != term.NoColorSupport {
			fwd += cfg.icon.size + padding
			fmt.Println(term.Fwd(fwd) + v + term.Reset)
		} else {
			fmt.Println(strings.Repeat(" ", int(fwd)) + v + term.Reset)
		}
	}
	lines += len(ss)
}

func printErr(label string, err error) {
	printLine(label, term.DarkYellow+"Failed "+term.Gray+"("+err.Error()+term.Gray+")")
}

func printTimeout(label string) {
	printLine(label, term.DarkYellow+"Timed out")
}

func printNetInfo(host string, port uint16) {
	var ip string
	if net.ParseIP(host) == nil {
		ips, err := net.LookupIP(host)
		if err == nil {
			ip = ips[0].String()
		}
	} else {
		ip = host
		host = ""
	}
	if host != "" {
		printLine("Host", cfg.host)
		if host != cfg.host {
			printLine("SRV", host)
		}
	}
	if ip != "" {
		printLine("IP", ip)
	}
	if port == 0 {
		port = cfg.port
	}
	if cfg.bedrock.enabled {
		port = cfg.bedrock.port
	}
	if port != 0 {
		printLine("Port", port)
	}
	if cfg.crossplay {
		printLine("Bedrock port", cfg.bedrock.port)
	}
}

func printMotd(s string) {
	ss := strings.Split(s, "\n")
	for i, s := range ss {
		ss[i] = term.TrimSpace(s)
	}
	if len(ss) > 1 {
		n := [2]int{utf8.RuneCountInString(term.RemoveCsi(ss[0])), utf8.RuneCountInString(term.RemoveCsi(ss[1]))}
		i := 0
		if n[1] < n[0] {
			i = 1
		}
		j := (i + 1) % 2
		ss[i] = strings.Repeat(" ", (n[j]-n[i])/2) + ss[i]
	}
	printLine("MOTD", strings.Join(ss, "\n"))
}

func printLatency(latency time.Duration) {
	ms := latency.Milliseconds()
	var c string
	switch {
	case ms < 50:
		c = term.Green
	case ms < 100:
		c = term.Yellow
	default:
		c = term.Red
	}
	printLine("Ping", fmt.Sprint(c, ms, " ms"))
}

func printPlayers(online, max int, sample []string) {
	s := fmt.Sprintf("%v"+term.Gray+"/"+term.Reset+"%v", online, max)
	for _, v := range sample {
		s += "\n" + mc.LegacyTextAnsi(v)
	}
	printLine("Players", s)
}

func printStatus(status *mc.StatusResponse) {
	if cfg.icon.enabled {
		printIcon(status.Icon)
	}

	printMotd(status.Motd.Ansi())

	printLatency(status.Latency)

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
			s = fmt.Sprintf("%v "+term.Gray+"(%v)", protoVerName, status.Version.Protocol)
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
		printLine("Prevents chat reports", term.Green+"Yes")
	}

	if len(status.Forge.Mods) > 0 {
		mods := make([]string, 0, len(status.Forge.Mods))
		for _, m := range status.Forge.Mods {
			mods = append(mods, m.Name+" "+term.Gray+m.Version)
		}
		printLine("Mods", strings.Join(mods, "\n"))
	}
}

func printBedrock(status mcpe.StatusResponse) {
	printLine("Name", mcpe.LegacyTextAnsi(status.Name))
	printLine("Level", mcpe.LegacyTextAnsi(status.Level))
	printLatency(status.Latency)
	printLine("Version", fmt.Sprintf("%v "+term.Gray+"(%v)", status.Version.Name, status.Version.Protocol))
	printPlayers(status.Players.Online, status.Players.Max, nil)
	printLine("Edition", status.Edition)
	printLine("Game Mode", fmt.Sprintf("%v "+term.Gray+"(%v)", status.GameMode.Name, status.GameMode.ID))
}

func printQuery(query mc.QueryResponse) {
	prev := lines
	if !cfg.status {
		printMotd(mc.LegacyTextAnsi(query.Motd))
		printLatency(query.Latency)
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
		printLine("Query", term.Green+"Enabled")
	}
}

func printResult[T any](result result[T], label string, fn func(T), failed string) {
	if result.success {
		fn(result.v)
	} else {
		if failed != "" {
			printLine(label, failed)
		} else if result.err != nil {
			printErr(label, result.err)
		} else {
			printTimeout(label)
		}
	}
}

func printResults(results *results) {
	host, port := cfg.host, cfg.port

	if cfg.icon.enabled && (!cfg.status || !results.status.success) {
		printIcon(nil)
	}

	if cfg.status {
		result := results.status
		if !result.success {
			cfg.status = false
		}
		s := "Status"
		if results.bedrock.success {
			s = "Java"
		}
		printResult(results.status, s, func(status mc.StatusResponse) {
			host, port = status.Host, status.Port
			printStatus(&status)
		}, term.Red+"Offline")
	}

	if cfg.crossplay && !results.status.success && results.bedrock.success {
		cfg.bedrock.enabled = true
		cfg.crossplay = false
	}
	if cfg.bedrock.enabled {
		printResult(results.bedrock, "Bedrock", func(status mcpe.StatusResponse) {
			port = cfg.bedrock.port
			printBedrock(status)
		}, term.Red+"Offline")
	}

	if cfg.query.enabled {
		result := results.query
		printResult(result, "Query", func(query mc.QueryResponse) {
			port = query.Port
			printQuery(query)
		}, term.Red+"Disabled")
	}

	if cfg.crossplay {
		printResult(results.bedrock, "Crossplay", func(status mcpe.StatusResponse) {
			printLine("Crossplay", term.Green+"Yes")
		}, term.Red+"No")
		if !results.bedrock.success {
			cfg.crossplay = false
		}
	}

	printNetInfo(host, port)

	if cfg.blocked {
		printResult(results.blocked, "Blocked", func(blocked string) {
			printLine("Blocked", formatBool(blocked == "", "No", fmt.Sprintf("Yes %v(%v)", term.Gray, blocked)))
		}, "")
	}

	if cfg.cracked {
		printResult(results.cracked, "Cracked", func(crackedWhitelisted crackedWhitelisted) {
			printLine("Cracked", formatBool(crackedWhitelisted.cracked, term.Reset+"Yes", term.Reset+"No"))
			if crackedWhitelisted.cracked {
				printLine("Whitelist", formatBool(!crackedWhitelisted.whitelisted, "Off", "On"))
			}
		}, "")
	}

	if cfg.rcon.enabled {
		printResult(results.rcon, "RCON", func(enabled bool) {
			printLine("RCON", formatBool(!enabled, "Disabled", "Enabled"))
		}, "")
	}

	if cfg.palette {
		printPalette()
	}

	if cfg.icon.enabled && term.ColorSupport != term.NoColorSupport && lines < int(iconHeight())+1 {
		fmt.Print(strings.Repeat("\n", int(iconHeight())-lines+1))
	} else {
		fmt.Print("\n")
	}
}

func printPalette() {
	const codes = "0123456789abcdef"
	var b strings.Builder
	b.WriteRune('\n')
	if cfg.icon.enabled {
		b.WriteString(term.Fwd(cfg.icon.size + padding))
	}
	for _, code := range codes[:len(codes)/2] {
		b.WriteString(term.Bg(mc.ParseColor(code)))
		b.WriteString("   ")
	}
	b.WriteString(term.Reset)
	b.WriteRune('\n')
	if cfg.icon.enabled {
		b.WriteString(term.Fwd(cfg.icon.size + padding))
	}
	for _, code := range codes[len(codes)/2:] {
		b.WriteString(term.Bg(mc.ParseColor(code)))
		b.WriteString("   ")
	}
	lines += 3
	if cfg.bedrock.enabled {
		const codes = "ghijmnpqstuv"
		b.WriteString(term.Reset)
		b.WriteRune('\n')
		if cfg.icon.enabled {
			b.WriteString(term.Fwd(cfg.icon.size + padding))
		}
		for _, code := range codes {
			b.WriteString(term.Bg(mcpe.ParseColor(code)))
			b.WriteString("  ")
		}
		lines++
	}
	b.WriteString(term.Reset)
	b.WriteRune('\n')
	fmt.Print(b.String())
}

func formatBool(bool bool, t, f string) string {
	var s string
	if bool {
		s = term.Green + t
	} else {
		s = term.Red + f
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
