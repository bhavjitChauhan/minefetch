package mc

import (
	"bytes"
	"cmp"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type StatusResponse struct {
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
	Favicon Icon
	// https://github.com/Aizistral-Studios/No-Chat-Reports/wiki/How-to-Get-Safe-Server-Status
	PreventsChatReports bool
	Latency             time.Duration
}

// TODO: merge host and port into address parameter

func Status(host string, port uint16, ver int32) (status StatusResponse, err error) {
	conn, err := net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(int(port))))
	if err != nil {
		return
	}
	defer conn.Close()

	err = writeHandshake(conn, ver, host, port, intentStatus)
	if err != nil {
		log.Fatalln("Failed to write handshake:", err)
	}

	err = writeStatusRequest(conn)
	if err != nil {
		log.Fatalln("Failed to write status request:", err)
	}

	err = readStatusResponse(conn, &status)
	if err != nil {
		log.Fatalln("Failed to read status response:", err)
	}

	start := time.Now()
	err = writePingRequest(conn, start.Unix())
	if err != nil {
		log.Fatalln("Failed to write ping request:", err)
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

	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.TrimPrefix(s, "data:image/png;base64,")))
	img, err := png.Decode(r)

	*icon = Icon{img, s}

	return err
}

func writeStatusRequest(w io.Writer) error {
	err1 := writeVarInt(w, 1)    // Packet length
	err2 := writeVarInt(w, 0x00) // Packet ID
	return cmp.Or(err1, err2)
}

func readStatusResponse(r io.Reader, status *StatusResponse) error {
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
