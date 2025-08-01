package mc

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func writePingRequest(w io.Writer, t int64) error {
	buf := &bytes.Buffer{}
	err1 := writeVarInt(buf, statusPacketIdPingRequest)
	err2 := binary.Write(buf, binary.BigEndian, t)
	if err := cmp.Or(err1, err2); err != nil {
		return err
	}

	return writePacket(w, buf.Bytes())
}

func readPongResponse(r io.Reader, t0 int64) error {
	id, buf, err := readPacket(r)
	if err != nil {
		return err
	}
	if id != statusPacketIdPongResponse {
		return errors.New(fmt.Sprint("unexpected packet ID: ", id))
	}

	var t1 int64
	err = binary.Read(buf, binary.BigEndian, &t1)
	if err != nil {
		return errors.New("failed to read timestamp: " + err.Error())
	}
	if t1 != t0 {
		return errors.New(fmt.Sprint("unexpected timestamp: ", t1))
	}

	return nil
}
