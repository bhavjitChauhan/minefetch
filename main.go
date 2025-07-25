package main

import (
	"fmt"
	"log"
	"minefetch/internal/mc"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	log.SetFlags(0)

	host, port, ver, err := parseArgs()
	if err != nil {
		log.Fatalln("Failed to parse arguments:", err)
	}

	timeout := time.After(cfg.timeout)

	chStatus := make(chan result)
	go func() {
		status, err := mc.Status(host, port, ver)
		chStatus <- result{-1, status, err}
	}()

	chResults := make(chan result)
	startResults(host, port, ver, chResults)

	printStatus(chStatus, timeout, host, port)
	printResults(chResults, timeout)

	if !cfg.noPalette {
		printPalette()
	}

	if !cfg.noIcon && lines < int(iconHeight())+1 {
		fmt.Print(strings.Repeat("\n", int(iconHeight())-lines+1))
	} else {
		fmt.Print("\n")
	}
}

func startResults(host string, port uint16, ver int32, ch chan<- result) {
	if cfg.query {
		go func() {
			queryPort := cfg.queryPort
			if queryPort == 0 {
				queryPort = uint(port)
			}
			address := net.JoinHostPort(host, strconv.Itoa(int(queryPort)))
			query, err := mc.Query(address)
			ch <- result{idQuery, query, err}
		}()
	}

	if cfg.blocked {
		go func() {
			blocked, err := mc.IsBlocked(host)
			ch <- result{idBlocked, blocked, err}
		}()
	}

	if cfg.cracked {
		go func() {
			// TODO: use server protocol from status response?
			cracked, whitelisted, err := mc.IsCracked(host, port, ver)
			ch <- result{idCracked, [2]bool{cracked, whitelisted}, err}
		}()
	}

	if cfg.rcon {
		go func() {
			address := net.JoinHostPort(host, strconv.Itoa(int(cfg.rconPort)))
			enabled, _ := mc.IsRconEnabled(address)
			ch <- result{idRcon, enabled, nil}
		}()
	}
}
