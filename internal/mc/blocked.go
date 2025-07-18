package mc

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

func IsBlocked(host string) (blocked bool, err error) {
	resp, err := http.Get("https://sessionserver.mojang.com/blockedservers")
	if err != nil {
		return
	}
	if resp.StatusCode != 200 {
		err = errors.New("status not ok: " + resp.Status)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: encode host in ISO-8859-1

	host = strings.ToLower(host)

	h := sha1.Sum([]byte(host))
	blocked = strings.Contains(string(body), hex.EncodeToString(h[:]))
	if blocked {
		return
	}

	// TODO: add IP wildcard support

	host = "." + host
	// Skip TLD check
	i := strings.LastIndexByte(host, '.')
	if i == -1 {
		return
	}
	i = strings.LastIndexByte(host[:i], '.')
	for i >= 0 {
		h := sha1.Sum([]byte("*" + host[i:]))
		blocked = strings.Contains(string(body), hex.EncodeToString(h[:]))
		if blocked {
			return
		}
		i = strings.LastIndexByte(host[:i], '.')
	}

	return
}
