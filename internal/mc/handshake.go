package mc

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"io"
)

type intent int32

const (
	intentStatus intent = iota + 1
	intentLogin
	intentTransfer
)

func writeHandshake(w io.Writer, ver int32, host string, port uint16, intent intent) error {
	buf := &bytes.Buffer{}
	err1 := writeVarInt(buf, 0x00)                    // Packet ID
	err2 := writeVarInt(buf, ver)                     // Protocol version
	err3 := writeString(buf, host)                    // Server address
	err4 := binary.Write(buf, binary.BigEndian, port) // Server port
	err5 := writeVarInt(buf, int32(intent))           // Intent
	if err := cmp.Or(err1, err2, err3, err4, err5); err != nil {
		return err
	}

	err1 = writeVarInt(w, int32(buf.Len()))
	_, err2 = w.Write(buf.Bytes())
	return cmp.Or(err1, err2)
}
