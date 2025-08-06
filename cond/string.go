package cond

import (
	"github.com/vela-public/onekit/cast"
	"github.com/vela-public/onekit/jsonkit"
	"github.com/vela-public/onekit/netkit"
	"path/filepath"
	"strconv"
	"strings"
)

type StringFSM struct {
	data string
	json *jsonkit.FastJSON
}

func (fsm *StringFSM) UnwrapJson() *jsonkit.FastJSON {
	if fsm.json == nil {
		obj := &jsonkit.FastJSON{}
		obj.ParseText(fsm.data)
		fsm.json = obj
	}
	return fsm.json
}

func (fsm *StringFSM) Getter(key string) string {
	sz := len(fsm.data)
	switch key {
	case "@len":
		return strconv.Itoa(len(fsm.data))
	case "@text":
		return fsm.data
	case "ext":
		return filepath.Ext(fsm.data)
	case "ipv4":
		return cast.ToString(netkit.Ipv4(fsm.data))
	case "ipv6":
		return cast.ToString(netkit.Ipv6(fsm.data))
	case "ip":
		return cast.ToString(netkit.Ipv4(fsm.data) || netkit.Ipv6(fsm.data))
	}

	if strings.HasPrefix(key, "json:") {
		k := strings.TrimPrefix(key, "json:")
		j := fsm.UnwrapJson()
		if j == nil {
			return ""
		}
		return j.Get(k).String()
	}

	if key[0] != '[' {
		return fsm.data
	}

	if key[sz-1] != ']' {
		return fsm.data
	}

	n := len(key)
	if n < 3 {
		return fsm.data
	}

	idx := strings.Index(key, ":")
	if idx < 0 {
		offset, err := cast.ToIntE(key[1 : n-1])
		if err != nil {
			return fsm.data
		}

		if offset >= 1 && offset <= sz {
			return string(fsm.data[offset-1])
		}

		return fsm.data
	}

	s := cast.ToInt(key[1:idx])
	e := cast.ToInt(key[idx+1 : n-1])
	if s > sz {
		return ""
	}

	if e == 0 || e > sz {
		return fsm.data[s:]
	}

	if s > e {
		return ""
	}
	return fsm.data[s:e]
}

func String(raw string) Lookup {
	fsm := &StringFSM{
		data: raw,
	}

	return fsm.Getter
}
