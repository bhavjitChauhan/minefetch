package mc

import (
	"bytes"
	"cmp"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
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

func WriteStatusRequest(w io.Writer) error {
	err1 := WriteVarInt(w, 1)    // Packet length
	err2 := WriteVarInt(w, 0x00) // Packet ID
	return cmp.Or(err1, err2)
}

func WritePingRequest(w io.Writer, t int64) error {
	buf := &bytes.Buffer{}
	err1 := WriteVarInt(buf, 0x01)
	err2 := binary.Write(buf, binary.BigEndian, t)
	if err := cmp.Or(err1, err2); err != nil {
		return err
	}

	err1 = WriteVarInt(w, int32(buf.Len()))
	_, err2 = w.Write(buf.Bytes())
	return cmp.Or(err1, err2)
}

type Status struct {
	Version struct {
		Name     string
		Protocol int32
	}
	EnforcesSecureChat bool
	Description        Text
	Players            struct {
		Max    int
		Online int
		Sample []struct {
			Id   string
			Name string
		}
	}
	Favicon string
	// https://github.com/Aizistral-Studios/No-Chat-Reports/wiki/How-to-Get-Safe-Server-Status
	PreventsChatReports bool
}

func ReadStatusResponse(r io.Reader, status *Status) error {
	n, err := ReadVarInt(r)
	if err != nil {
		return errors.New("failed to read length: " + err.Error())
	}

	buf := bytes.NewBuffer(make([]byte, n))
	_, err = io.ReadFull(r, buf.Bytes())
	if err != nil {
		return errors.New("failed to read: " + err.Error())
	}

	id, err := ReadVarInt(buf)
	if err != nil {
		return errors.New("failed to read packet ID: " + err.Error())
	}
	if id != 0x00 {
		return errors.New(fmt.Sprint("unexpected packet ID: ", id))
	}

	s, err := ReadString(buf)
	if err != nil {
		return errors.New("failed to read string: " + err.Error())
	}

	err = json.Unmarshal([]byte(s), &status)
	if err != nil {
		return errors.New("failed to parse JSON: " + err.Error())
	}

	return nil
}

func ReadPongResponse(r io.Reader, t0 int64) error {
	n, err := ReadVarInt(r)
	if err != nil {
		return errors.New("failed to read length: " + err.Error())
	}

	buf := bytes.NewBuffer(make([]byte, n))
	_, err = io.ReadFull(r, buf.Bytes())
	if err != nil {
		return errors.New("failed to read: " + err.Error())
	}

	id, err := ReadVarInt(buf)
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
