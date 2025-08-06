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

// lookupHostPort resolves the address using the following rules:
//   - If address is an IP with port, return the IP and port
//   - If address is an IP with no port, return the IP and defPort
//   - If address is a host with port, return SRV host if it exists, or the address host, both with address port
//   - If address is a host with no port, return the SRV host and port if they exist, or the host and defPort
func lookupHostPort(address string, defPort uint16) (host string, port uint16) {
	var err error
	host, port, err = SplitHostPort(address)
	noPort := port == 0 || err != nil
	if noPort {
		host = address
		port = defPort
		err = nil
	}
	if net.ParseIP(host) != nil {
		return
	}
	_, addrs, err := net.LookupSRV("minecraft", "tcp", host)
	if err != nil || len(addrs) == 0 {
		return
	}
	host = addrs[0].Target
	if host[len(host)-1] == '.' {
		host = host[:len(host)-1]
	}
	if noPort {
		port = addrs[0].Port
	}
	return
}
