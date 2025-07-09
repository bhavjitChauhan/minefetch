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
	"slices"
	"strconv"
	"strings"
	"time"
)

func main() {
	log.SetFlags(0)
	logger := log.Default()

	if len(os.Args) < 2 {
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

		err1 := writeVarInt(buf, 0x00)                    // Packet ID
		err2 := writeVarInt(buf, 771)                     // Protocol version	(1.21.6)
		err3 := writeString(buf, host)                    // Server address
		err4 := binary.Write(buf, binary.BigEndian, port) // Server port
		err5 := writeVarInt(buf, 1)                       // Intent				(Status)
		if err := cmp.Or(err1, err2, err3, err4, err5); err != nil {
			logger.Fatalln("Failed to create handshake:", err)
		}

		err1 = writeVarInt(conn, int32(buf.Len()))
		_, err2 = conn.Write(buf.Bytes())
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to write handshake:", err)
		}
	}

	{
		// Status request packet
		err1 := writeVarInt(conn, 1)
		err2 := writeVarInt(conn, 0x00) // Packet ID
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to write status request:", err)
		}
	}

	var status status
	{
		n, err := readVarInt(conn)
		if err != nil {
			logger.Fatalln("Failed to read status response length:", err)
		}

		buf := bytes.NewBuffer(make([]byte, n))
		_, err = io.ReadFull(conn, buf.Bytes())
		if err != nil {
			logger.Fatalln("Failed to read status response:", err)
		}

		id, err := readVarInt(buf)
		if err != nil {
			logger.Fatalln("Failed to read status response packet ID:", err)
		}
		if id != 0x00 {
			logger.Fatalln("Recieved unexpected status response packet ID:", id)
		}

		s, err := readString(buf)
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

		err1 := writeVarInt(buf, 0x01)
		err2 := binary.Write(buf, binary.BigEndian, start.Unix())
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to create ping request:", err)
		}

		err1 = writeVarInt(conn, int32(buf.Len()))
		_, err2 = conn.Write(buf.Bytes())
		if err := cmp.Or(err1, err2); err != nil {
			logger.Fatalln("Failed to write ping request:", err)
		}

		n, err := readVarInt(conn)
		if err != nil {
			logger.Fatalln("Failed to read pong response length:", err)
		}

		buf = bytes.NewBuffer(make([]byte, n))
		_, err = io.ReadFull(conn, buf.Bytes())
		if err != nil {
			logger.Fatalln("Failed to read pong response:", err)
		}

		id, err := readVarInt(buf)
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

	err = printIcon(&status.Favicon)
	if err != nil {
		logger.Fatalln("Failed to print icon:", err)
	}

	{
		type line struct {
			key string
			val any
		}
		var lines []line

		lines = append(lines, line{"Host", host}, line{"Port", port})
		if net.ParseIP(host) == nil {
			ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
			lines = append(lines, line{"IP", ip})
		}
		lines = append(lines,
			line{"MOTD", status.Description.ansi()},
			line{"Ping", latency.Milliseconds()})
		players := fmt.Sprintf("%v/%v", status.Players.Online, status.Players.Max)
		for _, v := range status.Players.Sample {
			players += "\n" + formatLegacy(v.Name)
		}
		lines = append(lines,
			line{"Players", players},
			line{"Version", fmt.Sprintf("%v (%v)", status.Version.Name, status.Version.Protocol)})

		fmt.Print(curUp(iconHeight-1), curBack(iconWidth))

		pad := len(slices.MaxFunc(lines, func(a, b line) int {
			return cmp.Compare(len(a.key), len(b.key))
		}).key) + 2
		for _, v := range lines {
			s := strings.Split(fmt.Sprint(v.val), "\n")
			fmt.Printf(curFwd(iconWidth+2)+"%-*v%v\n", pad, v.key+":", s[0])
			for _, v := range s[1:] {
				fmt.Println(curFwd(iconWidth+uint(pad)+2) + v)
			}
		}
		c := 0
		for _, line := range lines {
			if s, ok := line.val.(string); ok {
				c += strings.Count(s, "\n")
			}
		}
		if len(lines)+c < iconHeight+1 {
			fmt.Print(strings.Repeat("\n", iconHeight-len(lines)-c+1))
		}
	}
}
