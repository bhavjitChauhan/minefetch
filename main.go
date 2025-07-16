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

	host, port, err := parseArgs()
	if err != nil {
		log.Fatalln("Failed to parse arguments:", err)
	}

	chStatus := make(chan mc.StatusResponse)
	chQuery := make(chan mc.QueryResponse)
	chErr := make(chan error)

	go func() {
		status, err := mc.Status(host, port, mc.V1_21_7)
		if err != nil {
			chErr <- err
			return
		}
		chStatus <- status
	}()

	go func() {
		address := net.JoinHostPort(host, strconv.Itoa(int(port)))
		query, err := mc.Query(address)
		if err != nil {
			chErr <- err
			return
		}
		chQuery <- query
	}()

	var status mc.StatusResponse
	select {
	case status = <-chStatus:
	case err := <-chErr:
		log.Fatalln("Failed to get server status:", err)
	case <-time.After(time.Millisecond * 1000):
		log.Fatalln("The server took too long to respond.")
	}

	var query *mc.QueryResponse
	select {
	case q := <-chQuery:
		query = &q
	case err := <-chErr:
		log.Fatalln("Failed to query server:", err)
	case <-time.After(time.Millisecond * 100):
		break
	}

	printStatus(host, port, &status, query)
}
