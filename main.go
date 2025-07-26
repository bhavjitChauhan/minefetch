package main

import (
	"log"
	"minefetch/internal/mc"
	"net"
	"strconv"
	"time"
)

func main() {
	log.SetFlags(0)

	host, port, ver, err := parseArgs()
	if err != nil {
		log.Fatalln("Failed to parse arguments:", err)
	}

	ch := make(chan result)
	timeout := time.After(cfg.timeout)
	startResults(ch, host, port, ver)
	printResults(ch, timeout, host, port)
}

func startResults(ch chan<- result, host string, port uint16, ver int32) {
	if !cfg.noStatus {
		go func() {
			status, err := mc.Status(host, port, ver)
			ch <- result{idStatus, status, err}
		}()
	}

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
