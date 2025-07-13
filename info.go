package main

import (
	"fmt"
	"minefetch/internal/ansi"
	"minefetch/internal/mc"
	"net"
	"os"
	"strings"
	"time"
)

type infoEntry struct {
	key string
	val any
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

	var ip string
	if net.ParseIP(host) == nil {
		ip, _, _ = net.SplitHostPort(conn.RemoteAddr().String())
	}

	argHost, _, err := net.SplitHostPort(os.Args[1])
	if err != nil {
		argHost = os.Args[1]
	}

	entries = append(entries, infoEntry{"MOTD", strings.Join(desc, "\n")})
	entries = append(entries, infoEntry{"Ping", fmt.Sprint(latency.Milliseconds(), " ms")})
	entries = append(entries, infoEntry{"Version", mc.LegacyTextAnsi(status.Version.Name)})
	entries = append(entries, infoEntry{"Players", players})
	entries = append(entries, infoEntry{"Host", argHost})
	if ip != "" {
		entries = append(entries, infoEntry{"IP", ip})
	}
	entries = append(entries, infoEntry{"Port", port})
	if host != argHost {
		entries = append(entries, infoEntry{"SRV Record", host})
	}

	fmt.Print(ansi.Up(iconHeight-1), ansi.Back(iconWidth))

	for _, e := range entries {
		s := strings.Split(fmt.Sprint(e.val), "\n")
		fmt.Println(ansi.Fwd(iconWidth+2) + ansi.Bold + ansi.Blue + e.key + ansi.Reset + ": " + s[0])
		for _, v := range s[1:] {
			fmt.Println(ansi.Fwd(iconWidth+2+uint(len(e.key))+2) + v)
		}
		fmt.Print(ansi.Reset)
	}

	nl := 0
	for _, e := range entries {
		if s, ok := e.val.(string); ok {
			nl += strings.Count(s, "\n")
		}
	}
	if len(entries)+nl < iconHeight+1 {
		fmt.Print(strings.Repeat("\n", iconHeight-len(entries)-nl+1))
	}
}
