package main

import (
	"sync"
	"time"

	"github.com/bhavjitChauhan/minefetch/internal/mc"
	"github.com/bhavjitChauhan/minefetch/internal/mcpe"
)

type result[T any] struct {
	v       T
	err     error
	success bool
}

type crackedWhitelisted struct {
	cracked     bool
	whitelisted bool
}

type results struct {
	status  result[mc.StatusResponse]
	bedrock result[mcpe.StatusResponse]
	query   result[mc.QueryResponse]
	blocked result[string]
	cracked result[crackedWhitelisted]
	rcon    result[bool]
}

func getResults() *results {
	var results results
	var wg sync.WaitGroup

	if cfg.status {
		wg.Go(func() {
			address := cfg.host
			if cfg.port != 0 {
				address = mc.JoinHostPort(cfg.host, cfg.port)
			}
			status, err := mc.Status(address, cfg.proto)
			results.status = result[mc.StatusResponse]{status, err, err == nil}
		})
	}
	if cfg.bedrock.enabled || cfg.crossplay {
		wg.Go(func() {
			status, err := mcpe.Status(mc.JoinHostPort(cfg.host, cfg.bedrock.port))
			results.bedrock = result[mcpe.StatusResponse]{status, err, err == nil}
		})
	}
	if cfg.query.enabled {
		wg.Go(func() {
			address := cfg.host
			queryPort := cfg.query.port
			if queryPort == 0 {
				queryPort = cfg.port
			}
			if queryPort != 0 {
				address = mc.JoinHostPort(cfg.host, queryPort)
			}
			query, err := mc.Query(address)
			results.query = result[mc.QueryResponse]{query, err, err == nil}
		})
	}
	if cfg.blocked {
		wg.Go(func() {
			blocked, err := mc.IsBlocked(cfg.host)
			results.blocked = result[string]{blocked, err, err == nil}
		})
	}
	if cfg.cracked {
		wg.Go(func() {
			// TODO: use server protocol from status response?
			address := cfg.host
			if cfg.port != 0 {
				address = mc.JoinHostPort(cfg.host, cfg.port)
			}
			cracked, whitelisted, err := mc.IsCracked(address, cfg.proto)
			results.cracked = result[crackedWhitelisted]{crackedWhitelisted{cracked, whitelisted}, err, err == nil}
		})
	}
	if cfg.rcon.enabled {
		wg.Go(func() {
			enabled, _ := mc.IsRconEnabled(mc.JoinHostPort(cfg.host, cfg.rcon.port))
			results.rcon = result[bool]{enabled, nil, true}
		})
	}

	done := make(chan struct{})
	timeout := time.After(cfg.timeout)
	go func() {
		defer close(done)
		wg.Wait()
	}()
	select {
	case <-done:
	case <-timeout:
	}
	return &results
}
