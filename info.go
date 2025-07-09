package main

import (
	"cmp"
	"fmt"
	"minefetch/internal/ansi"
	"minefetch/internal/mc"
	"net"
	"slices"
	"strings"
	"time"
)

type infoEntry struct {
	key string
	val any
}

func printInfo(host string, port uint16, conn net.Conn, latency time.Duration, status *mc.Status) {
	var entries []infoEntry
	entries = append(entries, infoEntry{"Host", host}, infoEntry{"Port", port})
	if net.ParseIP(host) == nil {
		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		entries = append(entries, infoEntry{"IP", ip})
	}
	entries = append(entries,
		infoEntry{"MOTD", status.Description.Ansi()},
		infoEntry{"Ping", latency.Milliseconds()})
	players := fmt.Sprintf("%v/%v", status.Players.Online, status.Players.Max)
	for _, v := range status.Players.Sample {
		players += "\n" + mc.LegacyTextAnsi(v.Name)
	}
	entries = append(entries,
		infoEntry{"Players", players},
		infoEntry{"Version", fmt.Sprintf("%v (%v)", status.Version.Name, status.Version.Protocol)})

	fmt.Print(ansi.Up(iconHeight-1), ansi.Back(iconWidth))

	pad := len(slices.MaxFunc(entries, func(a, b infoEntry) int {
		return cmp.Compare(len(a.key), len(b.key))
	}).key) + 2
	for _, v := range entries {
		s := strings.Split(fmt.Sprint(v.val), "\n")
		fmt.Printf(ansi.Fwd(iconWidth+2)+"%-*v%v\n", pad, v.key+":", s[0])
		for _, v := range s[1:] {
			fmt.Println(ansi.Fwd(iconWidth+uint(pad)+2) + v)
		}
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
