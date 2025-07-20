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
	chStatusErr := make(chan error)
	go func() {
		status, err := mc.Status(host, port, mc.V1_21_7)
		if err != nil {
			chStatusErr <- err
			return
		}
		chStatus <- status
	}()

	chQuery := make(chan mc.QueryResponse)
	chQueryErr := make(chan error)
	go func() {
		address := net.JoinHostPort(host, strconv.Itoa(int(port)))
		query, err := mc.Query(address)
		if err != nil {
			chQueryErr <- err
			return
		}
		chQuery <- query
	}()

	chBlocked := make(chan string)
	chBlockedErr := make(chan error)
	go func() {
		blocked, err := mc.IsBlocked(host)
		if err != nil {
			chBlockedErr <- err
			return
		}
		chBlocked <- blocked
	}()

	type crackedData struct {
		cracked     bool
		whitelisted bool
	}
	chCracked := make(chan crackedData)
	chCrackedErr := make(chan error)
	go func() {
		// TODO: use server protocol from status response?
		cracked, whitelisted, err := mc.IsCracked(host, port, mc.V1_21_7)
		if err != nil {
			chCrackedErr <- err
			return
		}
		chCracked <- crackedData{cracked, whitelisted}
	}()

	var status mc.StatusResponse
	select {
	case status = <-chStatus:
	case err := <-chStatusErr:
		log.Fatalln("Failed to get server status:", err)
	case <-time.After(time.Second * 5):
		log.Fatalln("The server took too long to respond.")
	}

	err = printIcon(&status.Favicon)
	if err != nil {
		log.Fatalln("Failed to print icon:", err)
	}

	fmt.Print(ansi.Up(iconHeight-1) + ansi.Back(iconWidth))
	printStatus(host, port, &status)

	select {
	case query := <-chQuery:
		printQuery(&query)
	case err := <-chQueryErr:
		printInfo(info{"Query", ansi.DarkYellow + "Failed: " + err.Error()})
	case <-time.After(time.Second):
		printInfo(info{"Query", ansi.DarkYellow + "Timed out"})
	}

	select {
	case blocked := <-chBlocked:
		printInfo(info{"Blocked", formatBool(blocked == "", "No", fmt.Sprintf("Yes %v(%v)", ansi.Gray, blocked))})
	case err := <-chBlockedErr:
		printInfo(info{"Blocked", ansi.DarkYellow + "Failed: " + err.Error()})
	case <-time.After(time.Second):
		printInfo(info{"Blocked", ansi.DarkYellow + "Timed out"})
	}

	select {
	case crackedData := <-chCracked:
		printInfo(info{"Cracked", formatBool(crackedData.cracked, ansi.Reset+"Yes", ansi.Reset+"No")})
		if crackedData.cracked {
			printInfo(info{"Whitelist", formatBool(!crackedData.whitelisted, "Off", "On")})
		}
	case err := <-chCrackedErr:
		printInfo(info{"Cracked", ansi.DarkYellow + "Failed: " + err.Error()})
	case <-time.After(time.Second):
		printInfo(info{"Cracked", ansi.DarkYellow + "Timed out"})
	}

	printPalette()
	if lines < iconHeight+1 {
		fmt.Print(strings.Repeat("\n", iconHeight-lines+1))
	}
}
