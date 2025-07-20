package mc

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

// Determines if a server is in offline mode by attempting to login.
func IsCracked(host string, port uint16, ver int32) (cracked bool, whitelisted bool, err error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		return
	}
	defer conn.Close()

	err = writeHandshake(conn, ver, host, port, intentLogin)
	if err != nil {
		return
	}

	err = writeLoginStart(conn, "minefetch", uuid{})
	if err != nil {
		return
	}

	id, buf, err := readPacket(conn)
	if err != nil {
		return
	}

	if id == loginPacketIdDisconnect {
		var s string
		s, err = readString(buf)
		if err != nil {
			return
		}

		var v map[string]any
		err = json.Unmarshal([]byte(s), &v)
		if err != nil {
			err = errors.New("disconnected: " + s)
			return
		}
		if v, ok := v["translate"]; ok {
			if v == "multiplayer.disconnect.not_whitelisted" {
				cracked, whitelisted = true, true
			} else {
				err = fmt.Errorf("disconnected: %v", v)
			}
			return
		}

		t := normText(v, Text{})
		err = fmt.Errorf("disconnected: %v", t.Ansi())
		return
	}

	if id == loginPacketIdSetCompression {
		id, _, err = readCompressedPacket(conn)
		if err != nil {
			return
		}
	}

	cracked = id == loginPacketIdLoginSuccess
	return
}

func writeLoginStart(w io.Writer, user string, uuid uuid) error {
	buf := &bytes.Buffer{}
	err1 := writeVarInt(buf, loginPacketIdLoginStart)
	err2 := writeString(buf, user)
	err3 := binary.Write(buf, binary.BigEndian, uuid)
	if err := cmp.Or(err1, err2, err3); err != nil {
		return err
	}

	return writePacket(w, buf.Bytes())
}
