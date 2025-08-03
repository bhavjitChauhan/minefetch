package mc

import (
	"net"
	"strconv"
)

// SplitHostPort is like net.SplitHostPort, but with a uint16 port.
func SplitHostPort(address string) (host string, port uint16, err error) {
	host, s, err := net.SplitHostPort(address)
	if err != nil {
		return
	}
	int, err := strconv.Atoi(s)
	if err != nil {
		return
	}
	port = uint16(int)
	return
}

// JoinHostPort is like net.JoinHostPort, but with a uint16 port.
func JoinHostPort(host string, port uint16) string {
	return net.JoinHostPort(host, strconv.Itoa(int(port)))
}
