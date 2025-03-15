package cond

import (
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/netkit"
	"path/filepath"
	"strings"
)

func String(raw string) Lookup {
	size := len(raw)

	return func(key string) string { // * , ext , ipv4, ipv6 , [1,3]
		switch key {
		case "*":
			return raw
		case "ext":
			return filepath.Ext(raw)
		case "ipv4":
			return cast.ToString(netkit.Ipv4(raw))
		case "ipv6":
			return cast.ToString(netkit.Ipv6(raw))
		case "ip":
			return cast.ToString(netkit.Ipv4(raw) || netkit.Ipv6(raw))
		}

		n := len(key)
		if n < 3 {
			return raw
		}

		if key[0] != '[' {
			return raw
		}

		if key[n-1] != ']' {
			return raw
		}

		idx := strings.Index(key, ":")
		if idx < 0 {
			offset, err := cast.ToIntE(key[1 : n-1])
			if err != nil {
				return raw
			}

			if offset >= 1 && offset <= len(raw) {
				return string(raw[offset-1])
			}

			return raw
		}

		s := cast.ToInt(key[1:idx])
		e := cast.ToInt(key[idx+1 : n-1])
		if s > size {
			return ""
		}

		if e == 0 || e > size {
			return raw[s:]
		}

		if s > e {
			return ""
		}

		return raw[s:e]
	}
}
