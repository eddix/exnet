package exnet

import (
	"context"
	"net"
	"time"
)

// Dialer to dial and return an exnet.Conn
type Dialer struct {
	dialer *net.Dialer
}

// Dial works like net.Dial but return an exnet.Conn
func Dial(network, address string) (net.Conn, error) {
	d := &Dialer{}
	return d.Dial(network, address)
}

// DialTimeout works like net.DialTimeout but return an exnet.Conn
func DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	d := Dialer{dialer: &net.Dialer{Timeout: timeout}}
	return d.Dial(network, address)
}

// Underlying return underlying net.Dialer
func (d *Dialer) Underlying() *net.Dialer {
	return d.dialer
}

// Dial without context
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

// DialContext dial with context
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if d.dialer == nil {
		d.dialer = &net.Dialer{}
	}
	conn, err := d.dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}
	return &Conn{_conn: conn}, nil
}
