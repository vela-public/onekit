package netkit

import (
	"net"
	"strconv"
)

func Ipv4(addr string) bool {
	ip := net.ParseIP(addr)
	if ip == nil {
		return false
	}

	for i := 0; i < len(addr); i++ {
		if addr[i] == '.' {
			return true
		}
	}

	return false
}

func Ipv6(addr string) bool {
	ip := net.ParseIP(addr)
	if ip == nil {
		return false
	}

	for i := 0; i < len(addr); i++ {
		if addr[i] == ':' {
			return true
		}
	}

	return false
}

func IPPort(v string) bool {
	host, port, err := net.SplitHostPort(v)
	if err != nil {
		return false
	}

	if !Ipv4(host) && !Ipv6(host) {
		return false
	}

	p, err := strconv.Atoi(port)
	if err != nil {
		return false
	}

	if p > 0 && p < 65535 {
		return true
	}

	return false
}
