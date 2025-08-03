package mc

import (
	"bytes"
	"cmp"
	"compress/zlib"
	"io"
)

const handshakePacketId int32 = 0

const (
	statusPacketIdStatusRequest int32 = iota
	statusPacketIdPingRequest
)
const (
	statusPacketIdStatusResponse int32 = iota
	statusPacketIdPongResponse
)

const (
	loginPacketIdLoginStart int32 = iota
	loginPacketIdEncryptionResponse
	loginPacketIdLoginPluginResponse
	loginPacketIdLoginAcknowledged
	loginPacketIdCookieResponse
)
const (
	loginPacketIdDisconnect int32 = iota
	loginPacketIdEncryptionRequest
	loginPacketIdLoginSuccess
	loginPacketIdSetCompression
	loginPacketIdLoginPluginRequest
	loginPacketIdLoginCookieRequest
)

// https://minecraft.wiki/w/Java_Edition_protocol/Packets#Without_compression
func writePacket(w io.Writer, p []byte) error {
	buf := &bytes.Buffer{}
	err1 := writeVarInt(buf, int32(len(p)))
	_, err2 := buf.Write(p)
	if err := cmp.Or(err1, err2); err != nil {
		return err
	}

	_, err := w.Write(buf.Bytes())
	return err
}

func readPacket(r io.Reader) (id int32, buf *bytes.Buffer, err error) {
	n, err := readVarInt(r)
	if err != nil {
		return
	}

	buf = bytes.NewBuffer(make([]byte, n))
	_, err = io.ReadFull(r, buf.Bytes())
	if err != nil {
		return
	}

	id, err = readVarInt(buf)
	if err != nil {
		return
	}

	return
}

// https://minecraft.wiki/w/Java_Edition_protocol/Packets#With_compression
func readCompressedPacket(r io.Reader) (id int32, buf *bytes.Buffer, err error) {
	n, err := readVarInt(r)
	if err != nil {
		return
	}

	zbuf := bytes.NewBuffer(make([]byte, n))
	_, err = io.ReadFull(r, zbuf.Bytes())
	if err != nil {
		return
	}

	l, err := readVarInt(zbuf)
	if err != nil {
		return
	}

	if l == 0 {
		buf = zbuf
		id, err = readVarInt(buf)
		return
	}

	buf = bytes.NewBuffer(make([]byte, 0, l))
	var rc io.ReadCloser
	rc, err = zlib.NewReader(zbuf)
	if err != nil {
		return
	}
	_, err = io.Copy(buf, rc)
	if err != nil {
		return
	}
	err = rc.Close()
	if err != nil {
		return
	}

	id, err = readVarInt(buf)

	return
}
