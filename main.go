package main

import (
	"log"
	"minefetch/internal/mc"
	"minefetch/internal/mcpe"
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
	switch cfg.output {
	case "print":
		printResults(results)
	case "raw":
		printRawResults(results)
	}
}

type result struct {
	i       int
	v       any
	err     error
	timeout bool
}

type results [6]result

const (
	resultStatus = iota
	resultBedrockStatus
	resultQuery
	resultBlocked
	resultCracked
	resultRcon
)

func startResults(ch chan<- result) {
	if cfg.status {
		go func() {
			address := cfg.host
			if cfg.port != 0 {
				address = mc.JoinHostPort(cfg.host, cfg.port)
			}
			status, err := mc.Status(address, cfg.proto)
			ch <- result{resultStatus, status, err, false}
		}()
	}

	if cfg.bedrock || cfg.crossplay {
		go func() {
			status, err := mcpe.Status(mc.JoinHostPort(cfg.host, cfg.bedrockPort))
			ch <- result{resultBedrockStatus, status, err, false}
		}()
	}

	if cfg.query {
		go func() {
			address := cfg.host
			queryPort := cfg.queryPort
			if queryPort == 0 {
				queryPort = cfg.port
			}
			if queryPort != 0 {
				address = mc.JoinHostPort(cfg.host, queryPort)
			}
			query, err := mc.Query(address)
			ch <- result{resultQuery, query, err, false}
		}()
	}

	if cfg.blocked {
		go func() {
			blocked, err := mc.IsBlocked(cfg.host)
			ch <- result{resultBlocked, blocked, err, false}
		}()
	}

	if cfg.cracked {
		go func() {
			// TODO: use server protocol from status response?
			address := cfg.host
			if cfg.port != 0 {
				address = mc.JoinHostPort(cfg.host, cfg.port)
			}
			cracked, whitelisted, err := mc.IsCracked(address, cfg.proto)
			ch <- result{resultCracked, [2]bool{cracked, whitelisted}, err, false}
		}()
	}

	if cfg.rcon {
		go func() {
			enabled, _ := mc.IsRconEnabled(mc.JoinHostPort(cfg.host, cfg.rconPort))
			ch <- result{resultRcon, enabled, nil, false}
		}()
	}
}

func collectResults(ch <-chan result, timeout <-chan time.Time) results {
	var results results
	n := countBools(cfg.status, cfg.bedrock, cfg.crossplay, cfg.query, cfg.blocked, cfg.cracked, cfg.rcon)
	if n == 0 {
		log.Fatalln("Nothing to do!")
	}
	for ; n > 0; n-- {
		select {
		case result := <-ch:
			results[result.i] = result
		case <-timeout:
			n = 0
		}
	}
	for i, r := range results {
		if r.err != nil {
			results[i].v = nil
		} else if r.v == nil && r.err == nil {
			results[i].timeout = true
		}
	}
	return results
}

func countBools(bools ...bool) int {
	n := 0
	for _, b := range bools {
		if b {
			n++
		}
	}
	return n
}
