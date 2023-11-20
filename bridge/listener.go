package bridge

import (
	"net"

	"github.com/powerpuffpenguin/vnet"
	"github.com/powerpuffpenguin/vnet/reverse"
)

type Listener struct {
	*reverse.Listener
}

func (l Listener) Accept() (c net.Conn, e error) {
	c, e = l.Listener.Accept()
	if e != nil && e != vnet.ErrListenerClosed {
		e = Error{e: e}
	}
	return
}

type Error struct {
	e error
}

func (e Error) Error() string {
	return e.e.Error()
}
func (Error) Timeout() bool {
	return true
}

func (Error) Temporary() bool {
	return true
}
