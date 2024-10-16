package netkit

import (
	"fmt"
	"github.com/vela-public/onekit/libkit"
	"net"
	"sort"
	"strings"
)

type NULL struct{}

type IPMatch struct {
	Name string
	IP4  []IP4Range
	IP6  *IP6Range
}

func (ipm *IPMatch) File(path string) error {
	return libkit.ReadlineFunc(path, func(text string) (stop bool, e error) {
		err := ipm.Add(text)
		if err != nil {
			return true, err
		}

		return false, nil
	})
}

func (ipm *IPMatch) Match(v string) bool {
	ip := net.ParseIP(v)
	if ip4 := ip.To4(); ip4 != nil {
		return ipm.MatchIPv4(ip4)
	}

	if ip6 := ip.To16(); ip6 != nil {
		return ipm.MatchIPv6(v)
	}
	return false
}

func (ipm *IPMatch) MatchIPv6(ip6 string) bool {
	if ipm.IP6 == nil {
		return false
	}
	return ipm.IP6.MatchRaw(ip6)
}

func (ipm *IPMatch) MatchIPv4(ip net.IP) bool {
	n := len(ipm.IP4)
	if n == 0 {
		return false
	}

	ipNum := IP4ToUint32(ip)
	i := sort.Search(n, func(i int) bool {
		ip4 := ipm.IP4[i]
		return ip4.Start <= ipNum && ipNum <= ip4.End
	})

	if i < n && ipm.IP4[i].Start <= ipNum && ipNum <= ipm.IP4[i].End {
		return true
	}

	return false
}

func (ipm *IPMatch) InsertIPv4(s, e net.IP) error {
	si, ei := IP4ToUint32(s), IP4ToUint32(e)
	if si > ei {
		return fmt.Errorf("ipv4 range start greater")
	}

	ipm.IP4 = append(ipm.IP4, IP4Range{si, ei})
	sort.Slice(ipm.IP4, func(i, j int) bool {
		return ipm.IP4[i].Start < ipm.IP4[j].Start
	})
	return nil
}

func (ipm *IPMatch) Add(v string) error {
	if s := strings.Index(v, "-"); s != -1 {
		sip := net.ParseIP(v[:s])
		eip := net.ParseIP(v[s+1:])
		sp4 := sip.To4()
		ep4 := eip.To4()
		if sp4 != nil && ep4 != nil {
			return ipm.InsertIPv4(sp4, ep4)
		}

		return fmt.Errorf("not ipv4 range %v", v)
	}

	ip := net.ParseIP(v) // single ip
	if ip4 := ip.To4(); ip4 != nil {
		return ipm.InsertIPv4(ip4, ip4)
	}

	if ip6 := ip.To16(); ip6 != nil {
		ipm.IP6.Add(v, NULL{})
		return nil
	}

	ip, _, err := net.ParseCIDR(v) // cidr ip range
	if err != nil {
		return err
	}

	if ip4 := ip.To4(); ip4 != nil {
		return ipm.InsertIPv4(ip4, ip4)
	}

	if ip6 := ip.To16(); ip6 != nil {
		return ipm.IP6.Insert(v)
	}

	return fmt.Errorf("not ipv4 or ipv6 %v", v)
}
