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

func writeHandshake(w io.Writer, proto int32, host string, port uint16, intent intent) error {
	buf := &bytes.Buffer{}
	err1 := writeVarInt(buf, handshakePacketId)
	err2 := writeVarInt(buf, proto)
	err3 := writeString(buf, host)
	err4 := binary.Write(buf, binary.BigEndian, port)
	err5 := writeVarInt(buf, int32(intent))
	if err := cmp.Or(err1, err2, err3, err4, err5); err != nil {
		return err
	}

	return writePacket(w, buf.Bytes())
}
