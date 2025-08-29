package mc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// StatusResponse contains general server info provided by the [Server List Ping interface].
//
// Any string field except Icon and Uuid may contain legacy formatting.
//
// Version.Name may contain any third-party server software name.
// This field is only visible on vanilla clients when the reported Protocol is incompatible.
// Some servers use this field along with an invalid Protocol to display arbitrary information.
//
// Sample is sometimes used to display arbitrary information.
//
// PreventsChatReports is sent by servers using plugins or mods like [No Chat Reports].
//
// Icon is the raw encoded PNG data.
//
// [Server List Ping interface]: https://minecraft.wiki/w/Java_Edition_protocol/Server_List_Ping
// [No Chat Reports]: https://github.com/Aizistral-Studios/No-Chat-Reports/wiki/How-to-Get-Safe-Server-Status
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
	Icon                Icon `json:"favicon"`
	PreventsChatReports bool
	Forge               struct {
		Version uint8
		Mods    []mod
	}

	Host    string
	Port    uint16
	Latency time.Duration
	Raw     string
}

type mod struct {
	Name    string
	Version string
}

type statusResponse struct {
	StatusResponse
	ForgeData struct {
		Channels []struct {
			Res      string
			Version  string
			Required bool
		}
		Mods []struct {
			ModId     string
			ModMarker string
		}
		FmlNetworkVersion int
		D                 string
	}
	ModInfo struct {
		Type    string
		ModList []struct {
			ModId   string
			Version string
		}
	}
}

// Status attempts to get general server info using the [Server List Ping interface].
//
// This is the same interface used by the in-game server list.
//
// Most MOTD plugins will fall back to legacy formatting codes for proto versions before 1.16.
//
// It is possible for servers to disable the Server List Ping interface,
// so no response does not necessarily mean a server is offline.
//
// Some servers only respond to a second request.
// This may be a countermeasure against server scanners like [Copenheimer].
//
// [Server List Ping interface]: https://minecraft.wiki/w/Java_Edition_protocol/Server_List_Ping
// [Copenheimer]: https://2b2t.miraheze.org/wiki/Fifth_Column#Copenheimer
func Status(address string, proto int32) (status StatusResponse, err error) {
	host, port := lookupHostPort(address, 25565)
	status.Host = host
	status.Port = port

	address = JoinHostPort(host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return
	}
	defer conn.Close()

	err = writeHandshake(conn, proto, host, port, intentStatus)
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

type Icon []byte

func (icon *Icon) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}
	// len("data:image/png;base64,") = 22
	text = text[22:]
	*icon = make([]byte, len(text))
	_, err := base64.StdEncoding.Decode(*icon, text)
	return err
}

// https://minecraft.wiki/w/Java_Edition_protocol/Server_List_Ping#Status_Request
func writeStatusRequest(w io.Writer) error {
	buf := &bytes.Buffer{}
	err := writeVarInt(buf, statusPacketIdStatusRequest)
	if err != nil {
		return err
	}

	return writePacket(w, buf.Bytes())
}

// https://minecraft.wiki/w/Java_Edition_protocol/Server_List_Ping#Status_Response
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

	var raw statusResponse
	raw.Raw = s

	err = json.Unmarshal([]byte(s), &raw)
	if err != nil {
		return errors.New("failed to parse JSON: " + err.Error())
	}

	if len(raw.ForgeData.Mods) > 0 {
		raw.Forge.Version = uint8(raw.ForgeData.FmlNetworkVersion)
		raw.Forge.Mods = make([]mod, 0, len(raw.ForgeData.Mods))
		for _, m := range raw.ForgeData.Mods {
			raw.Forge.Mods = append(raw.Forge.Mods, mod{m.ModId, m.ModMarker})
		}
	} else if len(raw.ModInfo.ModList) > 0 {
		raw.Forge.Version = 1
		raw.Forge.Mods = make([]mod, 0, len(raw.ModInfo.ModList))
		for _, m := range raw.ModInfo.ModList {
			raw.Forge.Mods = append(raw.Forge.Mods, mod{m.ModId, m.Version})
		}
	}

	*status = raw.StatusResponse
	return nil
}
