package main

import (
	"fmt"
	"log"
	"minefetch/internal/ansi"
	"minefetch/internal/flag"
	"minefetch/internal/mc"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var cfg = struct {
	host        string
	port        uint16
	help        bool
	proto       int32
	timeout     time.Duration
	status      bool
	icon        bool
	iconSize    uint
	bedrock     bool
	bedrockPort uint16
	crossplay   bool
	query       bool
	queryPort   uint16
	blocked     bool
	cracked     bool
	rcon        bool
	rconPort    uint16
	palette     bool
	color       string
	output      string
	argHost     string
}{
	host:        "localhost",
	port:        25565,
	status:      true,
	icon:        true,
	bedrockPort: 19132,
	crossplay:   true,
	timeout:     time.Second,
	iconSize:    32,
	rconPort:    25575,
	palette:     true,
	color:       "auto",
	output:      "print",
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

func parseArgs() (err error) {
	proto := "1.21.8"
	flag.Var(&cfg.help, "help", 'h', cfg.help, "")
	flag.Var(&proto, "proto", 'p', proto, "Protocol version to use for requests.")
	flag.Var(&cfg.timeout, "timeout", 't', cfg.timeout, "Maximum time to wait for a response before timing out.")
	flag.Var(&cfg.status, "no-status", 0, cfg.status, "Don't get server info using the Server List Ping interface.")
	flag.Var(&cfg.icon, "no-icon", 0, cfg.icon, "Don't print the server icon.")
	flag.Var(&cfg.iconSize, "icon-size", 0, cfg.iconSize, "Icon size in pixels.")
	flag.Var(&cfg.bedrock, "bedrock", 'b', cfg.bedrock, "Get Bedrock server info.")
	flag.Var(&cfg.bedrockPort, "bedrock-port", 0, cfg.bedrockPort, "Bedrock server port.")
	flag.Var(&cfg.crossplay, "no-crossplay", 0, cfg.crossplay, "Don't check if a Bedrock server is running on the same host.")
	flag.Var(&cfg.query, "query", 'q', cfg.query, "Get server info using the query protocol.")
	flag.Var(&cfg.queryPort, "query-port", 0, "port", "Query protocol port.")
	flag.Var(&cfg.blocked, "blocked", 0, cfg.blocked, "Check the host against Mojang's blocklist.")
	flag.Var(&cfg.cracked, "cracked", 'c', cfg.cracked, "Attempt to login using an offline player.")
	flag.Var(&cfg.rcon, "rcon", 'r', cfg.rcon, "Check if the RCON protocol is enabled.")
	flag.Var(&cfg.rconPort, "rcon-port", 0, cfg.rconPort, "RCON protocol port.")
	flag.Var(&cfg.palette, "no-palette", 0, cfg.palette, "Print Minecraft's formatting code colors.")
	flag.Var(&cfg.color, "color", 0, cfg.color, "Override terminal color support detection. (0, 16, 256, true)")
	flag.Var(&cfg.output, "output", 'o', cfg.output, "Output format. (print, raw)")

	args, err := flag.Parse()
	if err != nil {
		return
	}

	if cfg.help {
		printHelp()
	}

	if cfg.bedrock {
		cfg.status = false
	}

	if !cfg.status {
		cfg.crossplay = false
	}

	if cfg.color != "auto" {
		switch cfg.color {
		case "0":
			ansi.ColorSupport = ansi.NoColorSupport
		case "16":
			ansi.ColorSupport = ansi.Color16Support
		case "256":
			ansi.ColorSupport = ansi.Color256Support
		// https://github.com/chalk/supports-color/blob/ae809ecabd5965d0685e7fc121efe98c47ad8724/index.js#L85-L87
		case "true", "16m", "full", "truecolor":
			ansi.ColorSupport = ansi.TrueColorSupport
		// https://bixense.com/clicolors
		case "", "always", "on":
			if ansi.ColorSupport == ansi.NoColorSupport {
				ansi.ColorSupport = ansi.Color16Support
			}
		case "never", "no":
			ansi.ColorSupport = ansi.NoColorSupport
		default:
			return fmt.Errorf("invalid colors: %v", cfg.color)
		}
	}
	if ansi.ColorSupport == ansi.NoColorSupport {
		ansi.NoColor()
		cfg.icon = false
		cfg.palette = false
	}

	var port uint16
	switch len(args) {
	case 0:
		cfg.argHost = cfg.host
	case 1:
		cfg.argHost, port, err = mc.SplitHostPort(args[0])
		if err != nil {
			err = nil
			cfg.argHost = args[0]
		}
	case 2:
		cfg.argHost = args[0]
		port, err = parseUint16(args[1])
		if err != nil {
			return
		}
	default:
		log.Print("Too many arguments.\n\n")
		printHelp()
	}

	cfg.host = cfg.argHost
	if !cfg.bedrock && net.ParseIP(cfg.host) == nil {
		_, addrs, err := net.LookupSRV("minecraft", "tcp", cfg.host)
		if err == nil && len(addrs) > 0 {
			cfg.host = strings.TrimSuffix(addrs[0].Target, ".")
			port = addrs[0].Port
		}
	}
	if port != 0 {
		cfg.port = port
		if cfg.bedrock {
			cfg.bedrockPort = port
		}
	}

	cfg.proto = parseFlagProto(proto)

	return
}

func parseFlagProto(proto string) int32 {
	v, ok := mc.VersionNameId[proto]
	if ok {
		return v
	}

	i, err := strconv.Atoi(proto)
	if err != nil {
		log.Fatalln("Failed to parse protocol version:", proto)
	}

	return int32(i)
}

func parseUint16(s string) (uint16, error) {
	int, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return uint16(int), nil
}
