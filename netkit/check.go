package netkit

import "net"

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
