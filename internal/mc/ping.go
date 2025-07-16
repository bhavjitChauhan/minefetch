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
	err1 := writeVarInt(buf, 0x01)
	err2 := binary.Write(buf, binary.BigEndian, t)
	if err := cmp.Or(err1, err2); err != nil {
		return err
	}

	err1 = writeVarInt(w, int32(buf.Len()))
	_, err2 = w.Write(buf.Bytes())
	return cmp.Or(err1, err2)
}

func readPongResponse(r io.Reader, t0 int64) error {
	n, err := readVarInt(r)
	if err != nil {
		return errors.New("failed to read length: " + err.Error())
	}

	buf := bytes.NewBuffer(make([]byte, n))
	_, err = io.ReadFull(r, buf.Bytes())
	if err != nil {
		return errors.New("failed to read: " + err.Error())
	}

	id, err := readVarInt(buf)
	if err != nil {
		return errors.New("failed to read packet ID: " + err.Error())
	}
	if id != 0x01 {
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
