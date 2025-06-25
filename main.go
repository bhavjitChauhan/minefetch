package main

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	log.SetFlags(0)
	logger := log.Default()

	if len(os.Args) < 2 {
		fmt.Println("Usage: mcping <host>")
		os.Exit(1)
	}

	var host string
	var port uint16
	{
		// TODO: add support for IPv6 addresses
		before, after, found := strings.Cut(os.Args[1], ":")
		host = before
		if net.ParseIP(before) == nil {
			_, addrs, err := net.LookupSRV("minecraft", "tcp", before)
			if err == nil && len(addrs) > 0 {
				host = strings.TrimSuffix(addrs[0].Target, ".")
				port = addrs[0].Port
			}
		}
		if found {
			i, err := strconv.Atoi(after)
			if err != nil {
				logger.Fatalln("Invalid port")
			}
			port = uint16(i)
		} else if port == 0 {
			port = 25565
		}
	}

	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(int(port)))
	if err != nil {
		logger.Fatalln("Failed to connect to server:", err)
	}
	defer conn.Close()

	// _, err = conn.Write([]byte{0xFE, 0x01}) // Legacy Server List Ping

	// TODO: record response latency

	// Handshake packet
	{
		buf := &bytes.Buffer{}

		err1 := WriteVarInt(buf, 0x00)        // Packet ID
		err2 := WriteVarInt(buf, -1)          // Protocol version
		err3 := WriteString(buf, host)        // Server address
		err4 := WriteUnsignedShort(buf, port) // Server port
		err5 := WriteVarInt(buf, 1)           // Intent
		if err := cmp.Or(err1, err2, err3, err4, err5); err != nil {
			logger.Fatalln("Failed to create handshake:", err)
		}

		err1 = WriteVarInt(conn, int32(buf.Len()))
		_, err2 = conn.Write(buf.Bytes())
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to send handshake:", err)
		}
	}

	{
		// Status request packet
		err1 := WriteVarInt(conn, 1)
		err2 := WriteVarInt(conn, 0x00) // Packet ID
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to send status request:", err)
		}
	}

	var status Status
	{
		n, err := ReadVarInt(conn)
		if err != nil {
			logger.Fatalln("Failed to read status response:", err)
		}

		buf := bytes.NewBuffer(make([]byte, n))
		_, err = io.ReadFull(conn, buf.Bytes())
		if err != nil {
			logger.Fatalln("Failed to read status response:", err)
		}

		b, err := buf.ReadByte()
		if err != nil {
			logger.Fatalln("Failed to read status response:", err)
		}
		if b != 0x00 {
			logger.Fatalln("Recieved unexpected packet ID:", b)
		}

		s, err := ReadString(buf)
		if err != nil {
			logger.Fatalln("Failed to read status string:", err)
		}

		// logger.Println(s)

		err = json.Unmarshal([]byte(s), &status)
		if err != nil {
			logger.Fatalln("Failed to parse status JSON:", err)
		}
	}

	const pad int = len("Description:") + 2 // "Description" is the longest field name
	fmt.Printf("%-*v%v\n", pad, "Host:", host)
	fmt.Printf("%-*v%v\n", pad, "Port:", port)
	if net.ParseIP(host) == nil {
		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		fmt.Printf("%-*v%v\n", pad, "IP:", ip)
	}
	fmt.Printf("%-*v", pad, "Description:")
	for i, v := range strings.Split(status.Description.raw, "\n") {
		if i == 0 {
			fmt.Print(v)
		} else {
			fmt.Print(strings.Repeat(" ", pad), v)
		}
		fmt.Print("\n")
	}
	fmt.Printf("%-*v%v/%v\n", pad, "Players:", status.Players.Online, status.Players.Max)
	if len(status.Players.Sample) > 0 {
		fmt.Print(strings.Repeat(" ", pad))
		for i, v := range status.Players.Sample {
			fmt.Print(v.Name)
			if i != len(status.Players.Sample)-1 {
				fmt.Print(", ")
			}
		}
		fmt.Print("\n")
	}
	fmt.Printf("%-*v%v (%v)\n", pad, "Version:", status.Version.Name, status.Version.Protocol)
}
