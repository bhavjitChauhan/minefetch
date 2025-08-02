package mc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"net"
	"strings"
	"time"
)

type StatusResponse struct {
	Version struct {
		Name     string
		Protocol int32
	}
	EnforcesSecureChat bool
	Motd               Text `json:"description"`
	Players            struct {
		Max    int
		Online int
		Sample []struct {
			Uuid string `json:"id"`
			Name string
		}
	}
	Icon Icon `json:"favicon"`
	// https://github.com/Aizistral-Studios/No-Chat-Reports/wiki/How-to-Get-Safe-Server-Status
	PreventsChatReports bool
	Latency             time.Duration
}

func Status(address string, ver int32) (status StatusResponse, err error) {
	host, port, err := SplitHostPort(address)
	if err != nil {
		return
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer conn.Close()

	err = writeHandshake(conn, ver, host, port, intentStatus)
	if err != nil {
		err = errors.New("Failed to write handshake: " + err.Error())
		return
	}

	err = writeStatusRequest(conn)
	if err != nil {
		err = errors.New("Failed to write status request: " + err.Error())
		return
	}

	err = readStatusResponse(conn, &status)
	if err != nil {
		err = errors.New("Failed to read status response: " + err.Error())
		return
	}

	start := time.Now()
	err = writePingRequest(conn, start.Unix())
	if err != nil {
		err = errors.New("Failed to write ping request: " + err.Error())
		return
	}

	readPongResponse(conn, start.Unix())

	status.Latency = time.Since(start)

	return
}

type Icon struct {
	image.Image
	Raw string
}

func (icon *Icon) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if s == "" {
		return nil
	}

	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.TrimPrefix(s, "data:image/png;base64,")))
	img, err := png.Decode(r)

	*icon = Icon{img, s}

	return err
}

func writeStatusRequest(w io.Writer) error {
	buf := &bytes.Buffer{}
	err := writeVarInt(buf, statusPacketIdStatusRequest)
	if err != nil {
		return err
	}

	return writePacket(w, buf.Bytes())
}

func readStatusResponse(r io.Reader, status *StatusResponse) error {
	id, buf, err := readPacket(r)
	if err != nil {
		return err
	}
	if id != 0x00 {
		return errors.New(fmt.Sprint("unexpected packet ID: ", id))
	}

	s, err := readString(buf)
	if err != nil {
		return errors.New("failed to read string: " + err.Error())
	}

	err = json.Unmarshal([]byte(s), &status)
	if err != nil {
		return errors.New("failed to parse JSON: " + err.Error())
	}

	return nil
}
