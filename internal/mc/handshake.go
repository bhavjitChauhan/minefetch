package mc

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"io"
)

type Intent int32

const (
	IntentStatus Intent = iota + 1
	IntentLogin
	IntentTransfer
)

func WriteHandshake(w io.Writer, ver int32, host string, port uint16, intent Intent) error {
	buf := &bytes.Buffer{}
	err1 := WriteVarInt(buf, 0x00)                    // Packet ID
	err2 := WriteVarInt(buf, ver)                     // Protocol version
	err3 := WriteString(buf, host)                    // Server address
	err4 := binary.Write(buf, binary.BigEndian, port) // Server port
	err5 := WriteVarInt(buf, int32(intent))           // Intent
	if err := cmp.Or(err1, err2, err3, err4, err5); err != nil {
		return err
	}

	err1 = WriteVarInt(w, int32(buf.Len()))
	_, err2 = w.Write(buf.Bytes())
	return cmp.Or(err1, err2)
}
