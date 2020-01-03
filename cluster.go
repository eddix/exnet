package exnet

import (
	"context"
	"net"
	"time"
)

var (
	defaultTCPKeepAlive       = true
	defaultTCPKeepAlivePeriod = 3 * time.Second
	defaultTCPLinger          = 1
	defaultTCPNoDelay         = true
)

// Cluster contain service info
type Cluster struct {
	DialTimeout   time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	AddressPicker AddressPicker

	_tcpKeepAlive       *bool
	_tcpKeepAlivePeriod *time.Duration
	_tcpLinger          *int
	_tcpNoDelay         *bool
}

// AddressPicker interface to get an address
type AddressPicker interface {
	Addr() net.Addr
}

// AddressPickerConcern interface to concern the address usage
type AddressPickerConcern interface {
	// Connected will be called when an address is connected
	Connected(net.Addr)
	// Disconnected will be called when an address is disconnected
	Disconnected(net.Addr)
	// Failure will be called when an address can't be connected
	Failure(net.Addr, error)
}

// Dial create an empty context and dial with it
func (c *Cluster) Dial(network, addr string) (net.Conn, error) {
	return c.DialContext(context.Background(), network, addr)
}

// DialContext dial and return an exnet.Conn, network and address is useless, we use
// AddressPicker to get one.
func (c *Cluster) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	addr := c.AddressPicker.Addr()
	dialer := &Dialer{
		dialer: &net.Dialer{
			Timeout: c.DialTimeout,
		},
	}
	conn, err := dialer.DialContext(ctx, addr.Network(), addr.String())
	// concern
	if apc, ok := c.AddressPicker.(AddressPickerConcern); ok {
		if err == nil {
			apc.Connected(addr)
		} else {
			apc.Failure(addr, err)
		}
	}
	if err != nil {
		return nil, err
	}
	// SetSockOpt for tcp connection
	switch ulconn := conn.(*Conn).Underlying().(type) {
	case *net.TCPConn:
		if err = c.tcpsetsockopt(ulconn); err != nil {
			conn.Close()
			return nil, err
		}
	}
	// SetDeadline
	conn.SetDeadline(time.Now().Add(c.DialTimeout).Add(c.WriteTimeout))
	return conn, nil
}

// TCPSetKeepAlive change keep-alive when setsockopt after dial.
// WARN: Keep-alive is enable by default, ensure you know what you are doing
//       when you call this function and change it
func (c *Cluster) TCPSetKeepAlive(keepAlive bool) {
	vcopy := keepAlive
	c._tcpKeepAlive = &vcopy
}

// TCPSetKeepAlivePeriod change keep-alive period when setsockopt after dial.
// WARN: Keep-alive period is enable 3 seconds, ensure you know what you are doing
//       when you call this function and change it
func (c *Cluster) TCPSetKeepAlivePeriod(d time.Duration) {
	vcopy := d
	c._tcpKeepAlivePeriod = &vcopy
}

// TCPSetLinger change linger when setsockopt after dial
// WARN: Linger is enable by default, ensure you know what you are doing
//       when you call this function and change it
func (c *Cluster) TCPSetLinger(linger int) {
	vcopy := linger
	c._tcpLinger = &vcopy
}

// TCPSetNoDelay set NoDelay when setsockopt after dial
// WARN: NoDelay is enabled by default, ensure you know what you are doing
//       when you call this function and change it
func (c *Cluster) TCPSetNoDelay(noDelay bool) {
	vcopy := noDelay
	c._tcpNoDelay = &vcopy
}

// tcpsetsockopt
func (c *Cluster) tcpsetsockopt(conn *net.TCPConn) error {
	var err error

	// Keep-Alive
	tcpKeepAlive := defaultTCPKeepAlive
	if c._tcpKeepAlive != nil {
		tcpKeepAlive = *c._tcpKeepAlive
	}
	err = conn.SetKeepAlive(tcpKeepAlive)
	if err != nil {
		return err
	}

	// Keep-Alive Period
	tcpKeepAlivePeriod := defaultTCPKeepAlivePeriod
	if c._tcpKeepAlivePeriod != nil {
		tcpKeepAlivePeriod = *c._tcpKeepAlivePeriod
	}
	err = conn.SetKeepAlivePeriod(tcpKeepAlivePeriod)
	if err != nil {
		return err
	}

	// Linger
	tcpLinger := defaultTCPLinger
	if c._tcpLinger != nil {
		tcpLinger = *c._tcpLinger
	}
	err = conn.SetLinger(tcpLinger)
	if err != nil {
		return err
	}

	// NoDelay
	tcpNoDelay := defaultTCPNoDelay
	if c._tcpNoDelay != nil {
		tcpNoDelay = *c._tcpNoDelay
	}
	err = conn.SetNoDelay(tcpNoDelay)
	if err != nil {
		return err
	}
	return nil
}
