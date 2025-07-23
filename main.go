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

	host, port, ver, err := parseArgs()
	if err != nil {
		log.Fatalln("Failed to parse arguments:", err)
	}

	timeout := time.After(flagTimeout)

	chStatus := make(chan mc.StatusResponse)
	chStatusErr := make(chan error)
	go func() {
		status, err := mc.Status(host, port, ver)
		if err != nil {
			chStatusErr <- err
			return
		}
		chStatus <- status
	}()

	var chQuery chan mc.QueryResponse
	var chQueryErr chan error
	if flagQuery {
		chQuery = make(chan mc.QueryResponse)
		chQueryErr = make(chan error)
		go func() {
			queryPort := flagQueryPort
			if queryPort == 0 {
				queryPort = uint(port)
			}
			address := net.JoinHostPort(host, strconv.Itoa(int(queryPort)))
			query, err := mc.Query(address)
			if err != nil {
				chQueryErr <- err
				return
			}
			chQuery <- query
		}()
	}

	var chBlocked chan string
	var chBlockedErr chan error
	if flagBlocked {
		chBlocked = make(chan string)
		chBlockedErr = make(chan error)
		go func() {
			blocked, err := mc.IsBlocked(host)
			if err != nil {
				chBlockedErr <- err
				return
			}
			chBlocked <- blocked
		}()
	}

	type crackedData struct {
		cracked     bool
		whitelisted bool
	}
	var chCracked chan crackedData
	var chCrackedErr chan error
	if flagCracked {
		chCracked = make(chan crackedData)
		chCrackedErr = make(chan error)
		go func() {
			// TODO: use server protocol from status response?
			cracked, whitelisted, err := mc.IsCracked(host, port, ver)
			if err != nil {
				chCrackedErr <- err
				return
			}
			chCracked <- crackedData{cracked, whitelisted}
		}()
	}

	var chRcon chan bool
	if flagRcon {
		chRcon = make(chan bool)
		go func() {
			address := net.JoinHostPort(host, strconv.Itoa(int(flagRconPort)))
			enabled, _ := mc.IsRconEnabled(address)
			chRcon <- enabled
		}()
	}

	var status mc.StatusResponse
	select {
	case status = <-chStatus:
	case err := <-chStatusErr:
		log.Fatalln("Failed to get server status:", err)
	case <-timeout:
		log.Fatalln("The server took too long to respond.")
	}

	if flagIcon {
		err = printIcon(&status.Favicon)
		if err != nil {
			log.Fatalln("Failed to print icon:", err)
		}
	}

	if flagIcon {
		fmt.Print(ansi.Up(iconHeight()-1) + ansi.Back(flagIconSize))
	}
	printStatus(host, port, &status)

	if flagQuery {
		select {
		case query := <-chQuery:
			printQuery(&query)
		case err := <-chQueryErr:
			printErr("Query", err)
		case <-timeout:
			printInfo(info{"Query", ansi.DarkYellow + "Timed out"})
		}
	}

	if flagBlocked {
		select {
		case blocked := <-chBlocked:
			printInfo(info{"Blocked", formatBool(blocked == "", "No", fmt.Sprintf("Yes %v(%v)", ansi.Gray, blocked))})
		case err := <-chBlockedErr:
			printErr("Blocked", err)
		case <-timeout:
			printInfo(info{"Blocked", ansi.DarkYellow + "Timed out"})
		}
	}

	if flagCracked {
		select {
		case crackedData := <-chCracked:
			printInfo(info{"Cracked", formatBool(crackedData.cracked, ansi.Reset+"Yes", ansi.Reset+"No")})
			if crackedData.cracked {
				printInfo(info{"Whitelist", formatBool(!crackedData.whitelisted, "Off", "On")})
			}
		case err := <-chCrackedErr:
			printErr("Cracked", err)
		case <-timeout:
			printInfo(info{"Cracked", ansi.DarkYellow + "Timed out"})
		}
	}

	if flagRcon {
		select {
		case enabled := <-chRcon:
			printInfo(info{"RCON", formatBool(!enabled, "Disabled", "Enabled")})
		case <-timeout:
			printInfo(info{"RCON", ansi.DarkYellow + "Timed out"})
		}
	}

	if flagPalette {
		printPalette()
	}

	if flagIcon && lines < int(iconHeight())+1 {
		fmt.Print(strings.Repeat("\n", int(iconHeight())-lines+1))
	} else {
		fmt.Print("\n")
	}
}
