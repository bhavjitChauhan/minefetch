package main

import (
	"fmt"
	"log"
	"minefetch/internal/flag"
	"minefetch/internal/mc"
	"minefetch/internal/term"
	"os"
	"strconv"
	"time"
)

var cfg = struct {
	help    bool
	host    string
	port    uint16
	timeout time.Duration
	proto   int32
	status  bool
	bedrock struct {
		enabled bool
		port    uint16
	}
	crossplay bool
	query     struct {
		enabled bool
		port    uint16
	}
	cracked bool
	blocked bool
	rcon    struct {
		enabled bool
		port    uint16
	}
	icon struct {
		enabled bool
		format  string
		size    uint
	}
	palette bool
	output  string
	color   string
}{
	host:      "localhost",
	status:    true,
	crossplay: true,
	bedrock: struct {
		enabled bool
		port    uint16
	}{port: 19132},
	timeout: time.Second,
	rcon: struct {
		enabled bool
		port    uint16
	}{port: 25575},
	icon: struct {
		enabled bool
		format  string
		size    uint
	}{enabled: true, size: 32},
	palette: true,
	output:  "print",
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
	flag.Var(&cfg.timeout, "timeout", 't', cfg.timeout, "Maximum time to wait for a response before timing out.")
	flag.Var(&proto, "proto", 'p', proto, "Protocol version to use for requests.")
	flag.Var(&cfg.status, "no-status", 0, cfg.status, "Don't get server info using the Server List Ping interface.")
	flag.Var(&cfg.bedrock.enabled, "bedrock", 'b', cfg.bedrock.enabled, "Get Bedrock server info.")
	flag.Var(&cfg.bedrock.port, "bedrock-port", 0, cfg.bedrock.port, "Bedrock server port.")
	flag.Var(&cfg.crossplay, "no-crossplay", 0, cfg.crossplay, "Don't check if a Bedrock server is running on the same host.")
	flag.Var(&cfg.query.enabled, "query", 'q', cfg.query.enabled, "Get server info using the query protocol.")
	flag.Var(&cfg.query.port, "query-port", 0, "auto", "Query protocol port.")
	flag.Var(&cfg.blocked, "blocked", 0, cfg.blocked, "Check the host against Mojang's blocklist.")
	flag.Var(&cfg.cracked, "cracked", 'c', cfg.cracked, "Attempt to login using an offline player.")
	flag.Var(&cfg.rcon.enabled, "rcon", 'r', cfg.rcon.enabled, "Check if the RCON protocol is enabled.")
	flag.Var(&cfg.rcon.port, "rcon-port", 0, cfg.rcon.port, "RCON protocol port.")
	flag.Var(&cfg.icon.enabled, "no-icon", 0, cfg.icon.enabled, "Don't print the server icon.")
	flag.Var(&cfg.icon.format, "icon", 0, "auto", "Icon print format. (sixel, half)")
	flag.Var(&cfg.icon.size, "icon-size", 0, cfg.icon.size, "Icon size in pixels.")
	flag.Var(&cfg.palette, "no-palette", 0, cfg.palette, "Print Minecraft's formatting code colors.")
	flag.Var(&cfg.output, "output", 'o', cfg.output, "Output format. (print, raw)")
	flag.Var(&cfg.color, "color", 0, "auto", "Override terminal color support detection. (0, 16, 256, true)")

	args, err := flag.Parse()
	if err != nil {
		return
	}

	if cfg.help {
		printHelp()
	}

	if cfg.bedrock.enabled {
		cfg.status = false
		cfg.query.enabled = false
		cfg.cracked = false
		cfg.rcon.enabled = false
	}

	if !cfg.status {
		cfg.crossplay = false
	}

	if cfg.color != "" {
		switch cfg.color {
		case "0":
			term.ColorSupport = term.NoColorSupport
		case "16":
			term.ColorSupport = term.Color16Support
		case "256":
			term.ColorSupport = term.Color256Support
		// https://github.com/chalk/supports-color/blob/ae809ecabd5965d0685e7fc121efe98c47ad8724/index.js#L85-L87
		case "true", "16m", "full", "truecolor":
			term.ColorSupport = term.TrueColorSupport
		// https://bixense.com/clicolors
		case "", "always", "on":
			if term.ColorSupport == term.NoColorSupport {
				term.ColorSupport = term.Color16Support
			}
		case "never", "no":
			term.ColorSupport = term.NoColorSupport
		default:
			return fmt.Errorf("invalid colors: %v", cfg.color)
		}
	}
	if term.ColorSupport == term.NoColorSupport {
		term.NoColor()
		cfg.palette = false
	}

	if cfg.icon.format != "" {
		if cfg.icon.format != "sixel" && cfg.icon.format != "half" {
			return fmt.Errorf("invalid icon type: %v", cfg.icon.format)
		}
	} else {
		if term.ColorSupport != term.NoColorSupport {
			cfg.icon.format = "half"
		} else {
			cfg.icon.format = "shade"
		}
	}

	var port uint16
	switch len(args) {
	case 0:
		// break
	case 1:
		cfg.host, port, err = mc.SplitHostPort(args[0])
		if err != nil {
			err = nil
			cfg.host = args[0]
		}
	case 2:
		cfg.host = args[0]
		port, err = parseUint16(args[1])
		if err != nil {
			return
		}
	default:
		log.Print("Too many arguments.\n\n")
		printHelp()
	}

	if port != 0 {
		cfg.port = port
		if cfg.bedrock.enabled {
			cfg.bedrock.port = port
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
