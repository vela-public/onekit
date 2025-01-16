package ssckit

import tun "github.com/vela-ssoc/vela-tunnel"

type NotifierType interface {
	Connected() error
	Disconnect(err error)
	Reconnected(addr *tun.Address)
	Shutdown(err error)
}

type Notifier struct {
	this *Application
}

func (n *Notifier) OnConnect(func() error) {

}

func (n *Notifier) Disconnect(err error) {

}

func (n *Notifier) Reconnected(addr *tun.Address) {

}

func (n *Notifier) Shutdown(err error) {

}
