package httpkit

import (
	"github.com/vela-public/onekit/cast"
	"net"
	"strings"
)

func Port(v uint16) func(r *RawHttp) {
	return func(r *RawHttp) {
		r.Port = cast.ToString(v)
	}
}

func Scheme(v string) func(r *RawHttp) {
	return func(r *RawHttp) {
		if v == "https" {
			r.TLS = true
		}
		r.Scheme = v
	}
}

func Peer(peer string) func(r *RawHttp) {
	return func(r *RawHttp) {
		if len(peer) == 0 {
			return
		}

		host, port, err := net.SplitHostPort(peer)
		if err != nil {
			r.Peer = peer
		}
		r.Port = port
		r.Peer = strings.TrimSpace(host)
	}

}

func Conn(netConn net.Conn) func(r *RawHttp) {
	return func(r *RawHttp) {
		r.NetConn = netConn
	}
}
