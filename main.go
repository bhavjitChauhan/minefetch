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

	conn, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		log.Fatalln("Failed to connect to server:", err)
	}
	defer conn.Close()

	err = mc.WriteHandshake(conn, mc.V1_21_7, host, port, mc.IntentStatus)
	if err != nil {
		log.Fatalln("Failed to write handshake:", err)
	}

	err = mc.WriteStatusRequest(conn)
	if err != nil {
		log.Fatalln("Failed to write status request:", err)
	}

	var status mc.Status
	err = mc.ReadStatusResponse(conn, &status)
	if err != nil {
		log.Fatalln("Failed to read status response:", err)
	}

	start := time.Now()
	err = mc.WritePingRequest(conn, start.Unix())
	if err != nil {
		log.Fatalln("Failed to write ping request:", err)
	}

	err = mc.ReadPongResponse(conn, start.Unix())
	if err != nil {
		log.Fatalln("Failed to read pong response:", err)
	}

	latency := time.Since(start)

	err = printIcon(&status.Favicon)
	if err != nil {
		log.Fatalln("Failed to print icon:", err)
	}

	printInfo(host, port, conn, latency, &status)
}
