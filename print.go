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
	"time"
	"unicode/utf8"
)

const (
	idStatus = iota
	idQuery
	idBlocked
	idCracked
	idRcon
)

const padding = 2

var lines = 0

type result struct {
	id  int
	v   any
	err error
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

func printStatus(status *mc.StatusResponse, host string, port uint16) {
	if !cfg.noIcon {
		err := printIcon(&status.Icon)
		if err != nil {
			log.Fatalln("Failed to print icon:", err)
		}
		fmt.Print(ansi.Up(iconHeight()-1) + ansi.Back(cfg.iconSize))
	}

	{
		ss := strings.Split(status.Motd.Ansi(), "\n")
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
		s := fmt.Sprintf("%v"+ansi.Gray+"/"+ansi.Reset+"%v", status.Players.Online, status.Players.Max)
		for _, v := range status.Players.Sample {
			s += "\n" + mc.LegacyTextAnsi(v.Name)
		}
		printData("Players", s)
	}

	{
		ip := argHost
		if net.ParseIP(argHost) == nil {
			ips, err := net.LookupIP(host)
			if err == nil {
				ip = ips[0].String()
			}
		}
		if argHost != ip {
			printData("Host", argHost)
		}
		if host != argHost {
			printData("SRV", host)
		}
		if ip != "" {
			printData("IP", ip)
		}
	}

	printData("Port", port)

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
	default:
		printTimeout(label)
	}
}

func printResults(ch <-chan result, timeout <-chan time.Time, host string, port uint16) {
	var results [5]result

	n := boolInt(!cfg.noStatus) + boolInt(cfg.query) + boolInt(cfg.blocked) + boolInt(cfg.cracked) + boolInt(cfg.rcon)
	if n == 0 {
		log.Fatalln("Nothing to do!")
	}
	for ; n > 0; n-- {
		select {
		case result := <-ch:
			results[result.id] = result
		case <-timeout:
			n = 0
		}
	}

	if !cfg.noStatus {
		result := results[idStatus]
		if result.err != nil {
			cfg.noIcon = true
		}
		printResult(results[idStatus], "Status", func(status mc.StatusResponse) {
			printStatus(&status, host, port)
		})
	}

	if cfg.query {
		printResult(results[idQuery], "Query", func(query mc.QueryResponse) {
			printQuery(&query)
		})
	}

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
