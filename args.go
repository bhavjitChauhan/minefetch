package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

var argHost, argPort string

var (
	flagHelp    = false
	flagIcon    = true
	flagQuery   = true
	flagBlocked = true
	flagCracked = false
	flagPalette = true
)

func printHelp(flagsHelp string) {
	fmt.Print("Usage: minefetch <address>\n", flagsHelp)
}

func parseArgs() (host string, port uint16, err error) {
	var args, flags []string

	for _, arg := range os.Args[1:] {
		if arg[0] == '-' {
			flags = append(flags, arg)
		} else {
			args = append(args, arg)
		}
	}

	var fs flag.FlagSet
	// TODO: add usage
	fs.BoolVar(&flagHelp, "help", flagHelp, "(-h)")
	fs.BoolVar(&flagIcon, "icon", flagIcon, "(-i)")
	fs.BoolVar(&flagQuery, "query", flagQuery, "(-q)")
	fs.BoolVar(&flagBlocked, "blocked", flagBlocked, "(-b)")
	fs.BoolVar(&flagCracked, "cracked", flagCracked, "(-c)")
	fs.BoolVar(&flagPalette, "palette", flagPalette, "(-p)")

	var flagsHelp string
	{
		buf := bytes.NewBufferString("Flags:\n")
		fs.SetOutput(buf)
		fs.PrintDefaults()
		flagsHelp = buf.String()
	}

	if len(args) != 1 {
		printHelp(flagsHelp)
		os.Exit(0)
	}

	fs.BoolVar(&flagHelp, "h", flagHelp, "")
	fs.BoolVar(&flagIcon, "i", flagIcon, "")
	fs.BoolVar(&flagQuery, "q", flagQuery, "")
	fs.BoolVar(&flagBlocked, "b", flagBlocked, "")
	fs.BoolVar(&flagCracked, "c", flagCracked, "")
	fs.BoolVar(&flagPalette, "p", flagPalette, "")
	fs.Parse(flags)

	if flagHelp {
		printHelp(flagsHelp)
		os.Exit(0)
	}

	if fs.NArg() != 0 {
		log.Panicln("Recieved argument in flags")
	}

	argHost, argPort, err = net.SplitHostPort(args[0])
	if err != nil {
		argHost = args[0]
	}
	host = argHost
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
