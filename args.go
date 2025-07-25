package main

import (
	"fmt"
	"log"
	"minefetch/internal/flag"
	"minefetch/internal/mc"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var argHost, argPort string

var cfg = struct {
	help      bool
	proto     string
	timeout   time.Duration
	noIcon    bool
	iconSize  uint
	query     bool
	queryPort uint
	blocked   bool
	cracked   bool
	rcon      bool
	rconPort  uint
	noPalette bool
}{
	proto:    "1.21.8",
	timeout:  time.Second,
	iconSize: 32,
	rconPort: 25575,
}

func printHelp() {
	fmt.Print(`Usage:
        minefetch
        minefetch [host] [port]
        minefetch [host[:port]]
Flags:
`)
	flag.Print()
	os.Exit(0)
}

func parseArgs() (host string, port uint16, ver int32, err error) {
	flag.Var(&cfg.help, "help", 'h', cfg.help, "")
	flag.Var(&cfg.proto, "proto", 'p', cfg.proto, "Protocol version to use for requests.")
	flag.Var(&cfg.timeout, "timeout", 't', cfg.timeout, "Maximum time to wait for a response before timing out.")
	flag.Var(&cfg.noIcon, "no-icon", 0, cfg.noIcon, "Print the server icon.")
	flag.Var(&cfg.iconSize, "icon-size", 0, cfg.iconSize, "Icon size in pixels.")
	flag.Var(&cfg.query, "query", 'q', cfg.query, "Attempt to communicate using the query protocol.")
	flag.Var(&cfg.queryPort, "query-port", 0, cfg.queryPort, "Port to use for the query protocol.")
	flag.Var(&cfg.blocked, "blocked", 'b', cfg.blocked, "Check the host against Mojang's blocklist.")
	flag.Var(&cfg.cracked, "cracked", 'c', cfg.cracked, "Attempt to login using an offline player.")
	flag.Var(&cfg.rcon, "rcon", 'r', cfg.rcon, "Check if the RCON protocol is enabled.")
	flag.Var(&cfg.rconPort, "rcon-port", 0, cfg.rconPort, "Port to use for the RCON protocol.")
	flag.Var(&cfg.noPalette, "no-palette", 0, cfg.noPalette, "Print Minecraft's formatting code colors")

	args, err := flag.Parse()
	if err != nil {
		log.Fatalln("Failed to parse flags:", err)
	}

	if cfg.help {
		printHelp()
	}

	switch len(args) {
	case 0:
		argHost = "localhost"
	case 1:
		argHost, argPort, err = net.SplitHostPort(args[0])
		if err != nil {
			err = nil
			argHost = args[0]
		}
	case 2:
		argHost = args[0]
		argPort = args[1]
	default:
		log.Print("Too many arguments.\n\n")
		printHelp()
	}

	host = argHost
	ver = parseFlagProto()

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

	return
}

func parseFlagProto() int32 {
	v, ok := mc.VersionNameId[cfg.proto]
	if ok {
		return v
	}

	i, err := strconv.Atoi(cfg.proto)
	if err != nil {
		log.Fatalln("Failed to parse protocol version:", cfg.proto)
	}

	return int32(i)
}
