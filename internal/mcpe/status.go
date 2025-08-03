package mcpe

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type StatusResponse struct {
	Edition string
	Name    string
	Version struct {
		Name     string
		Protocol int
	}
	Players struct {
		Max    int
		Online int
	}
	ID       string
	Level    string
	GameMode struct {
		Name string
		ID   int
	}
	Port struct {
		IPv4, IPv6 uint16
	}
}

// Status attempts to get general server info using the [RakNet protocol].
//
// This is the same interface used by the in-game server list.
//
// [RakNet protocol]: https://minecraft.wiki/w/RakNet
func Status(address string) (status StatusResponse, err error) {
	udpAddr, _ := net.ResolveUDPAddr("udp", address)
	conn, err := net.Dial("udp", udpAddr.String())
	if err != nil {
		return
	}
	err = writeUnconnectedPing(conn)
	if err != nil {
		return
	}
	status, err = readUnconnectedPong(conn)
	if err != nil {
		return
	}
	return
}

var magic = [16]byte{0x00, 0xff, 0xff, 0x00, 0xfe, 0xfe, 0xfe, 0xfe, 0xfd, 0xfd, 0xfd, 0xfd, 0x12, 0x34, 0x56, 0x78}

const packetIdUnconnectedPing byte = 0x01

// https://minecraft.wiki/w/RakNet#Unconnected_Ping
func writeUnconnectedPing(w io.Writer) error {
	buf := &bytes.Buffer{}
	err1 := buf.WriteByte(packetIdUnconnectedPing)
	err2 := binary.Write(buf, binary.BigEndian, time.Now().Unix())
	err3 := binary.Write(buf, binary.BigEndian, magic)
	err4 := binary.Write(buf, binary.BigEndian, int64(0))
	_, err5 := w.Write(buf.Bytes())
	return cmp.Or(err1, err2, err3, err4, err5)
}

// https://minecraft.wiki/w/RakNet#Unconnected_Pong
func readUnconnectedPong(r io.Reader) (status StatusResponse, err error) {
	br := bufio.NewReader(r)
	_, err = br.Discard(33) // packet ID + time + guid + magic
	if err != nil {
		return
	}

	var n uint16
	err = binary.Read(br, binary.BigEndian, &n)
	if err != nil {
		return
	}

	b := make([]byte, n)
	_, err = br.Read(b)
	if err != nil {
		return
	}
	if len(b) == 0 {
		err = errors.New("zero-length response")
		return
	}

	ss := strings.Split(string(b), ";")
	status.Edition = ss[0]
	status.Name = ss[1]
	status.Version.Protocol, err = strconv.Atoi(ss[2])
	if err != nil {
		return
	}
	status.Version.Name = ss[3]
	status.Players.Online, err = strconv.Atoi(ss[4])
	if err != nil {
		return
	}
	status.Players.Max, err = strconv.Atoi(ss[5])
	if err != nil {
		return
	}
	status.ID = ss[6]
	status.Level = ss[7]
	status.GameMode.Name = ss[8]
	if len(ss) == 9 {
		return
	}
	status.GameMode.ID, err = strconv.Atoi(ss[9])
	if err != nil {
		return
	}
	if len(ss) == 10 {
		return
	}
	ipv4Port, err := strconv.Atoi(ss[10])
	if err != nil {
		return
	}
	status.Port.IPv4 = uint16(ipv4Port)
	ipv6Port, err := strconv.Atoi(ss[11])
	if err != nil {
		return
	}
	status.Port.IPv6 = uint16(ipv6Port)

	return
}
