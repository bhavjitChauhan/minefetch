package mc

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// IsRconEnabled reports whether the [remote console] (RCON) is enabled on the server at address.
//
// An empty-password login request is made,
// and the validity of the response is used to determine whether RCON is enabled or not.
//
// [remote console]: https://minecraft.wiki/w/RCON
func IsRconEnabled(address string) (enabled bool, err error) {
	_, argPort, err := SplitHostPort(address)
	host, port := lookupHostPort(address, 25575)
	if err == nil {
		port = argPort
	}
	address = JoinHostPort(host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer conn.Close()

	err = writeRconPacket(conn, rconPacketTypeLoginRequest, "")
	if err != nil {
		return
	}

	_, _, _, err = readRconPacket(conn)
	if err != nil {
		return
	}

	enabled = true
	return
}

const (
	rconPacketTypeLoginRequest  int32 = 3
	rconPacketTypeLoginResponse int32 = 2
	rconPacketTypeCommand       int32 = 2
	rconPacketTypeMulti         int32 = 0
	rconPacketTypeFailed        int32 = -1
)

// https://minecraft.wiki/w/RCON#Packet_format
func writeRconPacket(w io.Writer, t int32, payload string) error {
	id := int32(time.Now().Unix())
	buf1 := &bytes.Buffer{}
	err1 := binary.Write(buf1, binary.LittleEndian, id)
	err2 := binary.Write(buf1, binary.LittleEndian, t)
	_, err3 := buf1.Write([]byte(payload))
	// Padding seems unnecessary
	_, err4 := buf1.Write([]byte{0, 0})
	buf2 := &bytes.Buffer{}
	err5 := binary.Write(buf2, binary.LittleEndian, int32(buf1.Len()))
	_, err6 := buf2.Write(buf1.Bytes())
	_, err7 := w.Write(buf2.Bytes())
	return cmp.Or(err1, err2, err3, err4, err5, err6, err7)
}

// Does not support multi-packet responses.
func readRconPacket(r io.Reader) (id int32, t int32, payload string, err error) {
	var n int32
	err = binary.Read(r, binary.LittleEndian, &n)
	if err != nil {
		return
	}
	if n < 9 {
		err = fmt.Errorf("invalid packet length: %v", n)
		return
	}

	b := make([]byte, n)
	_, err = io.ReadFull(r, b)
	if err != nil {
		return
	}

	buf := bytes.NewBuffer(b)
	err1 := binary.Read(buf, binary.LittleEndian, &id)
	err2 := binary.Read(buf, binary.LittleEndian, &t)
	payload, err3 := buf.ReadString(0)
	err = cmp.Or(err1, err2, err3)

	return
}
