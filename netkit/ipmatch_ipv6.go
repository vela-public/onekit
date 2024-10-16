package netkit

import (
	"github.com/gaissmai/bart"
	"net/netip"
)

type IP6Range struct {
	Map map[string]NULL
	Tab *bart.Table[NULL]
}

func (ip6 *IP6Range) MatchRaw(ip string) bool {
	if ip6.Map != nil {
		_, ok := ip6.Map[ip]
		return ok
	}
	addr, _ := netip.ParseAddr(ip)
	_, ok := ip6.Tab.Lookup(addr)
	return ok
}

func (ip6 *IP6Range) Add(ip string, v NULL) {
	if ip6.Map == nil {
		ip6.Map = make(map[string]NULL)
	}

	ip6.Map[ip] = v
}

func (ip6 *IP6Range) Insert(v string) error {
	if ip6.Tab == nil {
		ip6.Tab = &bart.Table[NULL]{}
	}

	pfx, err := netip.ParsePrefix(v)
	if err != nil {
		return err
	}
	ip6.Tab.Insert(pfx, NULL{})
	return nil
}
