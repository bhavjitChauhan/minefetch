package mc

import (
	"bufio"
	"bytes"
	"cmp"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// QueryResponse contains general server info provided by the [query protocol].
//
// [query protocol]: https://minecraft.wiki/w/Query
type QueryResponse struct {
	Motd string
	Game struct {
		Type, Id string
	}
	Version  string
	Software string
	Plugins  []string
	World    string
	Players  struct {
		Max    int
		Online int
		Sample []string
	}
	Port uint16
	Ip   net.IP

	Host      string
	QueryPort uint16
	Latency   time.Duration
	Raw       string
}

// Query attempts to get general server info using the [query protocol].
//
// It is very similar to Status, but may contain more players and additional information about plugins.
// However, the query protocol is not widely enabled by public servers.
//
// The query protocol encodes strings in ISO 8859-1.
// Query will convert all strings to UTF-8 to support legacy formatting codes.
//
// [query protocol]: https://minecraft.wiki/w/Query
func Query(address string) (query QueryResponse, err error) {
	host, port := lookupHostPort(address, 25565)
	query.Host = host
	query.QueryPort = port
	address = JoinHostPort(host, port)
	addr, _ := net.ResolveUDPAddr("udp", address)
	start := time.Now()

	conn, err := net.Dial("udp", addr.String())
	if err != nil {
		return
	}
	defer conn.Close()

	id := int32(time.Now().Unix()) & 0x0f0f0f0f
	err = writeQueryHandshake(conn, id)
	if err != nil {
		return
	}

	token, err := readQueryHandshake(conn, id)
	if err != nil {
		return
	}

	query.Latency = time.Since(start)

	err = writeQueryStatus(conn, id, token)
	if err != nil {
		return
	}

	query, err = readQueryStatus(conn, id)
	if err != nil {
		return
	}

	return
}

const queryMagic uint16 = 0xFEFD

type queryPacketType byte

const (
	queryPacketTypeHandshake queryPacketType = 9
	queryPacketTypeStat      queryPacketType = 0
)

// https://minecraft.wiki/w/Query#Client_to_Server_Packet_Format
func writeQueryPacket(w io.Writer, t queryPacketType, id int32, payload ...any) (err error) {
	buf := &bytes.Buffer{}
	err1 := binary.Write(buf, binary.BigEndian, queryMagic)
	err2 := binary.Write(buf, binary.BigEndian, t)
	err3 := binary.Write(buf, binary.BigEndian, id)
	if payload[0] != nil {
		for _, p := range payload {
			err = binary.Write(buf, binary.BigEndian, p)
			if err != nil {
				return
			}
		}
	}
	_, err4 := w.Write(buf.Bytes())
	err = cmp.Or(err1, err2, err3, err4)
	return
}

// https://minecraft.wiki/w/Query#Request
func writeQueryHandshake(w io.Writer, id int32) error {
	return writeQueryPacket(w, queryPacketTypeHandshake, id, nil)
}

// https://minecraft.wiki/w/Query#Request_3
func writeQueryStatus(w io.Writer, id, token int32) error {
	return writeQueryPacket(w, queryPacketTypeStat, id, token, int32(0))
}

// https://minecraft.wiki/w/Query#Server_to_Client_Packet_Format
func readQueryPacketHeader(r *bufio.Reader, t queryPacketType, id int32) (err error) {
	var st queryPacketType
	var sid int32

	err1 := binary.Read(r, binary.BigEndian, &st)
	err2 := binary.Read(r, binary.BigEndian, &sid)
	if err = cmp.Or(err1, err2); err != nil {
		return
	}
	if st != t {
		err = fmt.Errorf("expected packet type %v, got type: %v", t, st)
		return
	}
	if sid != id {
		err = fmt.Errorf("expected session id %v, got: %v", id, sid)
		return
	}

	return
}

// https://minecraft.wiki/w/Query#Response
func readQueryHandshake(r io.Reader, id int32) (token int32, err error) {
	br := bufio.NewReader(r)

	err = readQueryPacketHeader(br, queryPacketTypeHandshake, id)
	if err != nil {
		return
	}

	s, err := br.ReadString(0)
	if err != nil {
		return
	}

	i, err := strconv.Atoi(s[:len(s)-1])
	if err != nil {
		return
	}

	token = int32(i)
	return
}

// https://minecraft.wiki/w/Query#Response_3
func readQueryStatus(r io.Reader, id int32) (query QueryResponse, err error) {
	br := bufio.NewReader(r)

	err = readQueryPacketHeader(br, queryPacketTypeStat, id)
	if err != nil {
		return
	}

	// Padding
	n, err := br.Discard(11)
	if err != nil {
		return
	}
	if n != 11 {
		err = fmt.Errorf("expected 11 bytes, got: %v", n)
		return
	}

	b, err := br.Peek(br.Buffered())
	if err != nil {
		return
	}
	query.Raw = string(b)

	for {
		var k string
		k, err = br.ReadString(0)
		if err != nil {
			return
		}
		if k == string(byte(0)) {
			break
		}

		var b []byte
		b, err = br.ReadBytes(0)
		if err != nil {
			return
		}
		// Convert ISO 8859-1 to UTF-8
		buf := make([]rune, len(b))
		for i, b := range b {
			buf[i] = rune(b)
		}
		v := string(buf)

		k = k[:len(k)-1]
		v = v[:len(v)-1]

		switch k {
		case "hostname":
			query.Motd = v
		case "gametype":
			query.Game.Type = v
		case "game_id":
			query.Game.Id = v
		case "version":
			query.Version = v
		case "plugins":
			i := strings.Index(v, ": ")
			if i == -1 {
				query.Software = v
				break
			}
			query.Software = v[:i]
			if strings.TrimSpace(v[i+2:]) != "" {
				query.Plugins = strings.Split(v[i+2:], "; ")
			}
		case "map":
			query.World = v
		case "numplayers":
			query.Players.Online, err = strconv.Atoi(v)
			if err != nil {
				return
			}
		case "maxplayers":
			query.Players.Max, err = strconv.Atoi(v)
			if err != nil {
				return
			}
		case "hostport":
			var i int
			i, err = strconv.Atoi(v)
			if err != nil {
				return
			}
			query.Port = uint16(i)
		case "hostip":
			query.Ip = net.ParseIP(v)
		}
	}

	// Padding
	n, err = br.Discard(10)
	if err != nil {
		return
	}
	if n != 10 {
		err = fmt.Errorf("expected 10 bytes, got: %v", n)
		return
	}

	for {
		var player string
		player, err = br.ReadString(0)
		if err != nil {
			return
		}
		if player == string(byte(0)) {
			break
		}

		player = player[:len(player)-1]
		query.Players.Sample = append(query.Players.Sample, player)
	}

	return
}
