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
	idQuery = iota
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

type info struct {
	label string
	data  any
}

func printInfo(label string, data any) {
	s := strings.Split(fmt.Sprint(data), "\n")
	if flagIcon {
		fmt.Print(ansi.Fwd(flagIconSize + padding))
	}
	fmt.Println(ansi.Bold + ansi.Blue + label + ansi.Reset + ": " + s[0])
	for _, v := range s[1:] {
		fwd := uint(len(label)) + 2
		if flagIcon {
			fwd += flagIconSize + padding
		}
		fmt.Println(ansi.Fwd(fwd) + v)
	}
	fmt.Print(ansi.Reset)
	lines += len(s)
}

func printErr(label string, err error) {
	printInfo(label, ansi.DarkYellow+"Failed "+ansi.Gray+"("+err.Error()+ansi.Gray+")")
}

func printTimeout(label string) {
	printInfo(label, ansi.DarkYellow+"Timed out")
}

func printStatus(ch <-chan result, timeout <-chan time.Time, host string, port uint16) {
	var result result
	select {
	case result = <-ch:
	case <-timeout:
		log.Fatalln("The server took too long to respond.")
	}

	if result.err != nil {
		log.Fatalln("Failed to get server status:", result.err)
	}

	status, ok := result.v.(mc.StatusResponse)
	if !ok {
		log.Panicln("Unexpected result value for status:", result.v)
	}

	if flagIcon {
		err := printIcon(&status.Favicon)
		if err != nil {
			log.Fatalln("Failed to print icon:", err)
		}
	}

	if flagIcon {
		fmt.Print(ansi.Up(iconHeight()-1) + ansi.Back(flagIconSize))
	}

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
		protoVerName, ok := mc.VersionIdName[status.Version.Protocol]
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
		printInfo(i.label, i.data)
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
		printInfo(i.label, i.data)
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

func printResults(ch <-chan result, timeout <-chan time.Time) {
	var results [4]result

	n := boolInt(flagQuery) + boolInt(flagBlocked) + boolInt(flagCracked) + boolInt(flagRcon)
	for ; n > 0; n-- {
		select {
		case result := <-ch:
			results[result.id] = result
		case <-timeout:
			n = 0
		}
	}

	if flagQuery {
		printResult(results[idQuery], "Query", func(query mc.QueryResponse) {
			printQuery(&query)
		})
	}

	if flagBlocked {
		printResult(results[idBlocked], "Blocked", func(blocked string) {
			printInfo("Blocked", formatBool(blocked == "", "No", fmt.Sprintf("Yes %v(%v)", ansi.Gray, blocked)))
		})
	}

	if flagCracked {
		printResult(results[idCracked], "Cracked", func(crackedWhitelisted [2]bool) {
			printInfo("Cracked", formatBool(crackedWhitelisted[0], ansi.Reset+"Yes", ansi.Reset+"No"))
			if crackedWhitelisted[0] {
				printInfo("Whitelist", formatBool(!crackedWhitelisted[1], "Off", "On"))
			}
		})
	}

	if flagRcon {
		printResult(results[idRcon], "RCON", func(enabled bool) {
			printInfo("RCON", formatBool(!enabled, "Disabled", "Enabled"))
		})
	}
}

func printPalette() {
	const codes = "0123456789abcdef"
	fmt.Print("\n")
	if flagIcon {
		fmt.Print(ansi.Fwd(flagIconSize + padding))
	}
	for i, code := range codes {
		fmt.Print(ansi.Bg(mc.ParseColor(code)) + "   ")
		if (i + 1) == (len(codes) / 2) {
			fmt.Print(ansi.Reset + "\n")
			if flagIcon {
				fmt.Print(ansi.Fwd(flagIconSize + padding))
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
