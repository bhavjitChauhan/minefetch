package mc

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
)

// IsBlocked reports whether host is listed in Mojang's [blocked servers list].
//
// [blocked servers list]: https://github.com/sudofox/mojang-blocklist
func IsBlocked(host string) (selector string, err error) {
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
		return
	}

	// TODO: encode host in ISO-8859-1

	host = strings.ToLower(host)

	h := sha1.Sum([]byte(host))
	blocked := strings.Contains(string(body), hex.EncodeToString(h[:]))
	if blocked {
		selector = host
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
			selector = "*" + host[i:]
			return
		}
		i = strings.LastIndexByte(host[:i], '.')
	}

	return
}
