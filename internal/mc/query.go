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
}

func Query(address string) (status QueryResponse, err error) {
	conn, err := net.Dial("udp", address)
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

	err = writeQueryStatus(conn, id, token)
	if err != nil {
		return
	}

	status, err = readQueryStatus(conn, id)
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

func writeQueryHandshake(w io.Writer, id int32) error {
	return writeQueryPacket(w, queryPacketTypeHandshake, id, nil)
}

func writeQueryStatus(w io.Writer, id, token int32) error {
	return writeQueryPacket(w, queryPacketTypeStat, id, token, int32(0))
}

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

func readQueryStatus(r io.Reader, id int32) (status QueryResponse, err error) {
	br := bufio.NewReader(r)

	err = readQueryPacketHeader(br, queryPacketTypeStat, id)
	if err != nil {
		return
	}

	n, err := br.Discard(11)
	if err != nil {
		return
	}
	if n != 11 {
		err = fmt.Errorf("expected 11 bytes, got: %v", n)
		return
	}

	for {
		var k, v string
		k, err = br.ReadString(0)
		if err != nil {
			return
		}
		if k == string(byte(0)) {
			break
		}

		v, err = br.ReadString(0)
		if err != nil {
			return
		}

		k = k[:len(k)-1]
		v = v[:len(v)-1]

		switch k {
		case "hostname":
			status.Motd = v
		case "gametype":
			status.Game.Type = v
		case "game_id":
			status.Game.Id = v
		case "version":
			status.Version = v
		case "plugins":
			i := strings.Index(v, ": ")
			if i == -1 {
				status.Software = v
				break
			}
			status.Software = v[:i]
			if strings.TrimSpace(v[i+2:]) != "" {
				status.Plugins = strings.Split(v[i+2:], "; ")
			}
		case "map":
			status.World = v
		case "numplayers":
			status.Players.Online, err = strconv.Atoi(v)
			if err != nil {
				return
			}
		case "maxplayers":
			status.Players.Max, err = strconv.Atoi(v)
			if err != nil {
				return
			}
		case "hostport":
			var i int
			i, err = strconv.Atoi(v)
			if err != nil {
				return
			}
			status.Port = uint16(i)
		case "hostip":
			status.Ip = net.ParseIP(v)
		}
	}

	n, err = br.Discard(10)
	if err != nil {
		return
	}
	if n != 10 {
		err = fmt.Errorf("expected 10 bytes, got: %v", n)
		return
	}

	for {
		p, err1 := br.ReadString(0)
		if err = err1; err != nil {
			return
		}
		if p == string(byte(0)) {
			break
		}

		p = p[:len(p)-1]
		status.Players.Sample = append(status.Players.Sample, p)
	}

	return
}
