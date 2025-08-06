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
)

// IsCracked reports whether the server at address has online mode disabled.
//
// An unauthenticated login request is sent, and the response is used to determine the mode.
// Whitelist detection is not accurate as servers can customize the disconnect message.
//
// Note that login attempts are logged in the server console,
// and operators will see an unexpected disconnect message there.
func IsCracked(address string, proto int32) (cracked bool, whitelisted bool, err error) {
	host, port := lookupHostPort(address, 25565)

	address = JoinHostPort(host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer conn.Close()

	err = writeHandshake(conn, proto, host, uint16(port), intentLogin)
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
