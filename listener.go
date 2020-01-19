package exnet

import (
	"net"
)

// Listener listen and return a exnet.Conn
type Listener struct {
	_l net.Listener

	_acceptCallback func(net.Conn) error
}

var _ net.Listener = &Listener{}

// WithListener return an exnet.Listener with an underlying net.Listener,
// if l is already an exnet.Listener, return its-self.
func WithListener(l net.Listener) *Listener {
	if el, ok := l.(*Listener); ok {
		return el
	}
	return &Listener{_l: l}
}

// Listen works like net.Listen, but return an exnet.Listener
func Listen(network, addr string) (*Listener, error) {
	l, err := net.Listen(network, addr)
	if err != nil {
		return nil, err
	}
	return &Listener{_l: l}, nil
}

// Underlying return underlying net.Listener
func (l *Listener) Underlying() net.Listener {
	return l._l
}

// SetAcceptCallback add an accept callback to every new connection
// when it's accepted
func (l *Listener) SetAcceptCallback(f func(net.Conn) error) {
	l._acceptCallback = f
}

// Accept new connections, return an exnet.Conn
// if SetAcceptCallback called, accpet callback will call on new
// connection, if callback return error, accpet return error.
func (l *Listener) Accept() (net.Conn, error) {
	rwc, err := l._l.Accept()
	if err != nil {
		return nil, err
	}
	c := &Conn{_conn: rwc}
	if l._acceptCallback != nil {
		err = l._acceptCallback(c)
		if err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, nil
}

// Close underlying listener
func (l *Listener) Close() error { return l._l.Close() }

// Addr return underlying addr
func (l *Listener) Addr() net.Addr { return l._l.Addr() }
