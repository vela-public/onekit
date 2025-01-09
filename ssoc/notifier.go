package ssoc

import tun "github.com/vela-ssoc/vela-tunnel"

type Notifier struct {
	std *Standard
}

func (n *Notifier) Disconnect(err error) {

}

func (n *Notifier) Reconnected(addr *tun.Address) {

}

func (n *Notifier) Shutdown(err error) {

}
