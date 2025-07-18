package main

import (
	"fmt"
	"log"
	"minefetch/internal/ansi"
	"minefetch/internal/mc"
	"net"
	"strconv"
	"strings"
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
	chBlocked := make(chan bool)
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

	go func() {
		blocked, err := mc.IsBlocked(host)
		if err != nil {
			chErr <- err
			return
		}
		chBlocked <- blocked
	}()

	var status mc.StatusResponse
	select {
	case status = <-chStatus:
	case err := <-chErr:
		log.Fatalln("Failed to get server status:", err)
	case <-time.After(time.Second * 5):
		log.Fatalln("The server took too long to respond.")
	}

	err = printIcon(&status.Favicon)
	if err != nil {
		log.Fatalln("Failed to print icon:", err)
	}

	fmt.Print(ansi.Up(iconHeight-1) + ansi.Back(iconWidth))
	lines := 0
	lines += printStatus(host, port, &status)
	fmt.Print(ansi.Fwd(iconWidth + padding))

	var query *mc.QueryResponse
	select {
	case q := <-chQuery:
		query = &q
	case err := <-chErr:
		log.Fatalln("Failed to query server:", err)
	case <-time.After(time.Millisecond * 100):
		break
	}

	fmt.Print(ansi.Back(iconWidth + padding))
	lines += printQuery(query)

	var blocked bool
	select {
	case blocked = <-chBlocked:
	case err := <-chErr:
		log.Fatalln("Failed to check if the server is blocked:", err)
	case <-time.After(time.Millisecond * 100):
		break
	}
	lines += printInfo(info{"Blocked", formatBool(!blocked, "No", "Yes")})

	lines += printPalette()
	if lines < iconHeight+1 {
		fmt.Print(strings.Repeat("\n", iconHeight-lines+1))
	}
}
