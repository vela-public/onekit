package netkit

import (
	"net"
)

type IP4Range struct {
	Start uint32
	End   uint32
}

func (ip4 IP4Range) Match(v net.IP) bool {
	if ip4.Start > IP4ToUint32(v) || ip4.End < IP4ToUint32(v) {
		return false
	}
	return true
}
