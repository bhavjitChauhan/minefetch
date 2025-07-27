package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"minefetch/internal/ansi"
	"minefetch/internal/image/pngconfig"
	"minefetch/internal/mc"
	"net"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	idStatus = iota
	idQuery
	idBlocked
	idCracked
	idRcon
)

type results [5]result

const padding = 2

var lines = 0

type result struct {
	id      int
	v       any
	err     error
	timeout bool
}

func printData(label string, data any) {
	s := strings.Split(fmt.Sprint(data), "\n")
	if !cfg.noIcon {
		fmt.Print(ansi.Fwd(cfg.iconSize + padding))
	}
	fmt.Println(ansi.Bold + ansi.Blue + label + ansi.Reset + ": " + s[0])
	for _, v := range s[1:] {
		fwd := uint(len(label)) + 2
		if !cfg.noIcon {
			fwd += cfg.iconSize + padding
		}
		fmt.Println(ansi.Fwd(fwd) + v)
	}
	fmt.Print(ansi.Reset)
	lines += len(s)
}

func printErr(label string, err error) {
	printData(label, ansi.DarkYellow+"Failed "+ansi.Gray+"("+err.Error()+ansi.Gray+")")
}

func printTimeout(label string) {
	printData(label, ansi.DarkYellow+"Timed out")
}

func printAddrInfo() {
	ip := cfg.argHost
	if net.ParseIP(cfg.argHost) == nil {
		ips, err := net.LookupIP(cfg.host)
		if err == nil {
			ip = ips[0].String()
		}
	}
	if cfg.argHost != ip {
		printData("Host", cfg.argHost)
	}
	if cfg.host != cfg.argHost {
		printData("SRV", cfg.host)
	}
	if ip != "" {
		printData("IP", ip)
	}
	printData("Port", cfg.port)
}

func printMotd(motd mc.Text) {
	ss := strings.Split(motd.Ansi(), "\n")
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
	printData("MOTD", strings.Join(ss, "\n"))
}

func printPlayers(online, max int, sample []string) {
	s := fmt.Sprintf("%v"+ansi.Gray+"/"+ansi.Reset+"%v", online, max)
	for _, v := range sample {
		s += "\n" + mc.LegacyTextAnsi(v)
	}
	printData("Players", s)
}

func printStatus(status *mc.StatusResponse, host string, port uint16) {
	if !cfg.noIcon {
		err := printIcon(&status.Icon)
		if err != nil {
			log.Fatalln("Failed to print icon:", err)
		}
		fmt.Print(ansi.Up(iconHeight()-1) + ansi.Back(cfg.iconSize))
	}

	printMotd(status.Motd)

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
		printData("Ping", fmt.Sprint(c, ms, " ms"))
	}

	printData("Version", mc.LegacyTextAnsi(status.Version.Name))

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
		printData("Protocol", s)
	}

	if status.Icon.Image != nil {
		iconConfig, _ := pngconfig.DecodeConfig(base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.TrimPrefix(status.Icon.Raw, "data:image/png;base64,"))))
		interlaced := ""
		if iconConfig.Interlaced {
			interlaced = "Interlaced "
		}
		printData("Icon", fmt.Sprintf("%v%v-bit %v", interlaced, iconConfig.BitDepth, formatColorType(iconConfig.ColorType)))
	} else {
		printData("Icon", "Default")
	}

	printData("Secure chat", formatBool(!status.EnforcesSecureChat, "Not enforced", "Enforced"))

	if status.PreventsChatReports {
		printData("Prevents chat reports", ansi.Green+"Yes")

	}
}

func printQuery(query *mc.QueryResponse) {
	prev := lines
	if cfg.noStatus {
		printData("MOTD", mc.LegacyTextAnsi(query.Motd))
		// TODO: ping
		printData("Version", mc.LegacyTextAnsi(query.Version))
		printPlayers(query.Players.Online, query.Players.Max, query.Players.Sample)
	}
	if query != nil {
		if query.Software != "" {
			printData("Software", query.Software)
		}

		if len(query.Plugins) > 0 {
			printData("Plugins", strings.Join(query.Plugins, "\n"))
		}
	}
	if lines == prev {
		printData("Query", ansi.Green+"Enabled")
	}
}

func printResult[T any](result result, label string, fn func(T)) {
	switch {
	case result.err != nil:
		printErr(label, result.err)
	case result.v != nil:
		v, ok := result.v.(T)
		if !ok {
			log.Panicf("Unexpected result value for %v: %v", label, result.v)
		}
		fn(v)
	case result.timeout:
		printTimeout(label)
	default:
		log.Panicln("Unexpected result state:", result)
	}
}

func printResults(results results) {
	if !cfg.noStatus {
		result := results[idStatus]
		if result.err != nil || result.timeout {
			cfg.noStatus = true
			cfg.noIcon = true
		}
		printResult(results[idStatus], "Status", func(status mc.StatusResponse) {
			printStatus(&status, cfg.host, cfg.port)
		})
	}

	if cfg.query {
		result := results[idQuery]
		printResult(result, "Query", func(query mc.QueryResponse) {
			printQuery(&query)
		})
	}

	printAddrInfo()

	if cfg.blocked {
		printResult(results[idBlocked], "Blocked", func(blocked string) {
			printData("Blocked", formatBool(blocked == "", "No", fmt.Sprintf("Yes %v(%v)", ansi.Gray, blocked)))
		})
	}

	if cfg.cracked {
		printResult(results[idCracked], "Cracked", func(crackedWhitelisted [2]bool) {
			printData("Cracked", formatBool(crackedWhitelisted[0], ansi.Reset+"Yes", ansi.Reset+"No"))
			if crackedWhitelisted[0] {
				printData("Whitelist", formatBool(!crackedWhitelisted[1], "Off", "On"))
			}
		})
	}

	if cfg.rcon {
		printResult(results[idRcon], "RCON", func(enabled bool) {
			printData("RCON", formatBool(!enabled, "Disabled", "Enabled"))
		})
	}

	if !cfg.noPalette {
		printPalette()
	}

	if !cfg.noIcon && lines < int(iconHeight())+1 {
		fmt.Print(strings.Repeat("\n", int(iconHeight())-lines+1))
	} else {
		fmt.Print("\n")
	}
}

func printPalette() {
	const codes = "0123456789abcdef"
	fmt.Print("\n")
	if !cfg.noIcon {
		fmt.Print(ansi.Fwd(cfg.iconSize + padding))
	}
	for i, code := range codes {
		fmt.Print(ansi.Bg(mc.ParseColor(code)) + "   ")
		if (i + 1) == (len(codes) / 2) {
			fmt.Print(ansi.Reset + "\n")
			if !cfg.noIcon {
				fmt.Print(ansi.Fwd(cfg.iconSize + padding))
			}
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

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
