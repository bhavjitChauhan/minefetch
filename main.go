package main

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
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
		argHost, argPort, err := net.SplitHostPort(os.Args[1])
		if err != nil {
			host = os.Args[1]
		} else {
			host = argHost
		}
		if net.ParseIP(host) == nil {
			_, addrs, err := net.LookupSRV("minecraft", "tcp", host)
			if err == nil && len(addrs) > 0 {
				host = strings.TrimSuffix(addrs[0].Target, ".")
				port = addrs[0].Port
			}
		}
		if argPort != "" {
			i, err := strconv.Atoi(argPort)
			if err != nil {
				logger.Fatalln("Invalid port:", argPort)
			}
			port = uint16(i)
		} else if port == 0 {
			port = 25565
		}
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		logger.Fatalln("Failed to connect to server:", err)
	}
	defer conn.Close()

	// _, err = conn.Write([]byte{0xFE, 0x01}) // Legacy Server List Ping

	// Handshake packet
	{
		buf := &bytes.Buffer{}

		err1 := WriteVarInt(buf, 0x00)                    // Packet ID
		err2 := WriteVarInt(buf, -1)                      // Protocol version
		err3 := WriteString(buf, host)                    // Server address
		err4 := binary.Write(buf, binary.BigEndian, port) // Server port
		err5 := WriteVarInt(buf, 1)                       // Intent
		if err := cmp.Or(err1, err2, err3, err4, err5); err != nil {
			logger.Fatalln("Failed to create handshake:", err)
		}

		err1 = WriteVarInt(conn, int32(buf.Len()))
		_, err2 = conn.Write(buf.Bytes())
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to write handshake:", err)
		}
	}

	{
		// Status request packet
		err1 := WriteVarInt(conn, 1)
		err2 := WriteVarInt(conn, 0x00) // Packet ID
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to write status request:", err)
		}
	}

	var status Status
	{
		n, err := ReadVarInt(conn)
		if err != nil {
			logger.Fatalln("Failed to read status response length:", err)
		}

		buf := bytes.NewBuffer(make([]byte, n))
		_, err = io.ReadFull(conn, buf.Bytes())
		if err != nil {
			logger.Fatalln("Failed to read status response:", err)
		}

		id, err := ReadVarInt(buf)
		if err != nil {
			logger.Fatalln("Failed to read status response packet ID:", err)
		}
		if id != 0x00 {
			logger.Fatalln("Recieved unexpected status response packet ID:", id)
		}

		s, err := ReadString(buf)
		if err != nil {
			logger.Fatalln("Failed to parse status response string:", err)
		}

		// logger.Println(s)

		err = json.Unmarshal([]byte(s), &status)
		if err != nil {
			logger.Fatalln("Failed to parse status response JSON:", err)
		}
	}

	var latency time.Duration
	{
		start := time.Now()
		buf := &bytes.Buffer{}

		err1 := WriteVarInt(buf, 0x01)
		err2 := binary.Write(buf, binary.BigEndian, start.Unix())
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to create ping request:", err)
		}

		err1 = WriteVarInt(conn, int32(buf.Len()))
		_, err2 = conn.Write(buf.Bytes())
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to write ping request:", err)
		}

		n, err := ReadVarInt(conn)
		if err != nil {
			logger.Fatalln("Failed to read pong response length:", err)
		}

		buf = bytes.NewBuffer(make([]byte, n))
		_, err = io.ReadFull(conn, buf.Bytes())
		if err != nil {
			logger.Fatalln("Failed to read pong response:", err)
		}

		id, err := ReadVarInt(buf)
		if err != nil {
			logger.Fatalln("Failed to parse pong response packet ID:", err)
		}
		if id != 0x01 {
			logger.Fatalln("Recieved unexpected packet ID:", id)
		}

		var timestamp int64
		err = binary.Read(buf, binary.BigEndian, &timestamp)
		if err != nil {
			logger.Fatalln("Failed to parse pong response timestamp:", err)
		}
		if timestamp != start.Unix() {
			logger.Fatalln("Recieved unexpected pong response timestamp:", timestamp)
		}

		latency = time.Since(start)
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
	fmt.Printf("%-*v%v ms\n", pad, "Ping:", latency.Milliseconds())
	fmt.Printf("%-*v%v/%v\n", pad, "Players:", status.Players.Online, status.Players.Max)
	// TODO: handle colored player sample
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
