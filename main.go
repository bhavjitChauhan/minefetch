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

	err := parseArgs()
	if err != nil {
		log.Fatalln("Failed to parse arguments:", err)
	}

	ch := make(chan result)
	timeout := time.After(cfg.timeout)
	startResults(ch)
	results := collectResults(ch, timeout)
	printResults(results)
}

func startResults(ch chan<- result) {
	if !cfg.noStatus {
		go func() {
			status, err := mc.Status(cfg.host, cfg.port, cfg.proto)
			ch <- result{idStatus, status, err, false}
		}()
	}

	if cfg.query {
		go func() {
			queryPort := cfg.queryPort
			if queryPort == 0 {
				queryPort = uint(cfg.port)
			}
			address := net.JoinHostPort(cfg.host, strconv.Itoa(int(queryPort)))
			query, err := mc.Query(address)
			ch <- result{idQuery, query, err, false}
		}()
	}

	if cfg.blocked {
		go func() {
			blocked, err := mc.IsBlocked(cfg.host)
			ch <- result{idBlocked, blocked, err, false}
		}()
	}

	if cfg.cracked {
		go func() {
			// TODO: use server protocol from status response?
			cracked, whitelisted, err := mc.IsCracked(cfg.host, cfg.port, cfg.proto)
			ch <- result{idCracked, [2]bool{cracked, whitelisted}, err, false}
		}()
	}

	if cfg.rcon {
		go func() {
			address := net.JoinHostPort(cfg.host, strconv.Itoa(int(cfg.rconPort)))
			enabled, _ := mc.IsRconEnabled(address)
			ch <- result{idRcon, enabled, nil, false}
		}()
	}
}

func collectResults(ch <-chan result, timeout <-chan time.Time) results {
	var results results
	n := boolInt(!cfg.noStatus) + boolInt(cfg.query) + boolInt(cfg.blocked) + boolInt(cfg.cracked) + boolInt(cfg.rcon)
	if n == 0 {
		log.Fatalln("Nothing to do!")
	}
	for ; n > 0; n-- {
		select {
		case result := <-ch:
			results[result.id] = result
		case <-timeout:
			n = 0
		}
	}
	for i, r := range results {
		if r.err == nil && r.v == nil {
			results[i].timeout = true
		}
	}
	return results
}
