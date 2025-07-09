package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func parseArgs() (host string, port uint16, err error) {
	if len(os.Args) < 2 {
		fmt.Println("Usage: minefetch <host>")
		os.Exit(0)
	}

	argHost, argPort, err := net.SplitHostPort(os.Args[1])
	if err != nil {
		host = os.Args[1]
	} else {
		host = argHost
	}
	if net.ParseIP(host) == nil {
		_, addrs, err := net.LookupSRV("minecraft", "tcp", host)
		if err == nil && len(addrs) > 0 {
			host = strings.TrimSuffix(addrs[0].Target, ".")
			port = addrs[0].Port
		}
	}
	if argPort != "" {
		var i int
		i, err = strconv.Atoi(argPort)
		if err != nil {
			return
		}
		port = uint16(i)
	} else if port == 0 {
		port = 25565
	}
	return host, port, nil
}
