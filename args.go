package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"minefetch/internal/mc"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var argHost, argPort string

var (
	flagHelp          = false
	flagProto         = "1.21.8"
	flagTimeout       = time.Second
	flagIcon          = true
	flagIconSize uint = 32
	flagQuery         = true
	flagBlocked       = true
	flagCracked       = false
	flagPalette       = true
)

func printHelp(flagsHelp string) {
	fmt.Print(`Usage:	minefetch
	minefetch [host] [port]
	minefetch [host[:port]]
`, flagsHelp)
	os.Exit(0)
}

func parseArgs() (host string, port uint16, ver int32, err error) {
	var fs flag.FlagSet
	fs.BoolVar(&flagHelp, "help", flagHelp, "(-h)")
	fs.StringVar(&flagProto, "proto", flagProto, "Protocol version to use for requests. (-p)")
	fs.DurationVar(&flagTimeout, "timeout", flagTimeout, "Maximum time to wait for a response before timing out. (-t)")
	fs.BoolVar(&flagIcon, "icon", flagIcon, "Print the server icon. (-i)")
	fs.UintVar(&flagIconSize, "icon-size", flagIconSize, "Icon size in pixels.")
	fs.BoolVar(&flagQuery, "query", flagQuery, "Attempt to communicate using the query protocol. (-q)")
	fs.BoolVar(&flagBlocked, "blocked", flagBlocked, "Check the host against Mojang's blocklist. (-b)")
	fs.BoolVar(&flagCracked, "cracked", flagCracked, "Attempt to login using an offline player. (-c)")
	fs.BoolVar(&flagPalette, "palette", flagPalette, "Print Minecraft's formatting code colors")

	var flagsHelp string
	{
		buf := bytes.NewBufferString("Flags:\n")
		fs.SetOutput(buf)
		fs.PrintDefaults()
		flagsHelp = buf.String()
	}

	fs.BoolVar(&flagHelp, "h", flagHelp, "")
	fs.DurationVar(&flagTimeout, "t", flagTimeout, "")
	fs.StringVar(&flagProto, "p", flagProto, "")
	fs.BoolVar(&flagIcon, "i", flagIcon, "")
	fs.BoolVar(&flagQuery, "q", flagQuery, "")
	fs.BoolVar(&flagBlocked, "b", flagBlocked, "")
	fs.BoolVar(&flagCracked, "c", flagCracked, "")
	fs.Parse(os.Args[1:])

	if flagHelp {
		printHelp(flagsHelp)
	}

	switch fs.NArg() {
	case 0:
		argHost = "localhost"
	case 1:
		argHost, argPort, err = net.SplitHostPort(fs.Arg(0))
		if err != nil {
			err = nil
			argHost = fs.Arg(0)
		}
	case 2:
		argHost = fs.Arg(0)
		argPort = fs.Arg(1)
	default:
		log.Print("Too many arguments.\n\n")
		printHelp(flagsHelp)
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
	v, ok := mc.VersionNameId[flagProto]
	if ok {
		return v
	}

	i, err := strconv.Atoi(flagProto)
	if err != nil {
		log.Fatalln("Failed to parse protocol version:", flagProto)
	}

	return int32(i)
}
