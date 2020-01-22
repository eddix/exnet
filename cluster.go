package exnet

import (
	"context"
	"net"
	"sync/atomic"
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
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	AddressPicker AddressPicker
	connpool      ConnPool

	tcpKeepAlive       bool
	tcpKeepAlivePeriod time.Duration
	tcpLinger          int
	tcpNoDelay         bool

	// metrics
	metricDialDirect    int64
	metricDialPoolReuse int64
}

// ClusterConfig expose config for cluster
type ClusterConfig struct {
	// DialTimeout is the default timeout when call Dial
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	// connection pool settings
	PoolConfig   *ConnPoolConfig
	UseAsyncPool bool
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

// NewCluster create new cluster with config and default options.
func NewCluster(conf *ClusterConfig) *Cluster {
	c := &Cluster{
		DialTimeout:        conf.DialTimeout,
		ReadTimeout:        conf.ReadTimeout,
		WriteTimeout:       conf.WriteTimeout,
		AddressPicker:      nil,
		tcpKeepAlive:       defaultTCPKeepAlive,
		tcpKeepAlivePeriod: defaultTCPKeepAlivePeriod,
		tcpLinger:          defaultTCPLinger,
		tcpNoDelay:         defaultTCPNoDelay,
	}
	if conf.PoolConfig != nil {
		if conf.UseAsyncPool {
			c.connpool = NewAsyncConnPool(conf.PoolConfig)
		} else {
			c.connpool = NewSyncConnPool(conf.PoolConfig)
		}
	}
	return c
}

// Dial create an empty context and dial with it
func (c *Cluster) Dial(network, addr string) (net.Conn, error) {
	return c.DialContext(context.Background(), network, addr)
}

// DialContext dial and return an exnet.Conn, network and address is useless, we use
// AddressPicker to get one.
func (c *Cluster) DialContext(ctx context.Context, _, _ string) (net.Conn, error) {
	if c.connpool != nil {
		if conn := c.connpool.Get(); conn != nil {
			if c.resetDeadlines(conn) == nil {
				atomic.AddInt64(&c.metricDialPoolReuse, 1)
				return &Conn{_conn: conn, closer: c}, nil
			}
			_ = conn.Close()
		}
	}
	atomic.AddInt64(&c.metricDialDirect, 1)
	return c.dialContextDirect(ctx)
}

func (c *Cluster) dialContextDirect(ctx context.Context) (net.Conn, error) {
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
	switch ulconn := UnwrapConn(conn).(type) {
	case *net.TCPConn:
		if err = c.tcpsetsockopt(ulconn); err != nil {
			ulconn.Close()
			return nil, err
		}
	}
	err = c.resetDeadlines(conn)
	if err != nil {
		_ = UnwrapConn(conn).Close()
		return nil, err
	}
	return conn, nil
}

func (c *Cluster) resetDeadlines(conn net.Conn) error {
	var err error
	// TODO: Set REAL Deadline
	err = conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	if err != nil {
		return err
	}
	err = conn.SetDeadline(time.Now().Add(c.ReadTimeout).Add(c.WriteTimeout))
	if err != nil {
		return err
	}
	return nil
}

// Close conn closer
func (c *Cluster) Close(conn net.Conn) error {
	if exconn, ok := conn.(*Conn); ok {
		if exconn.err != nil {
			return UnwrapConn(conn).Close()
		}
	}
	if c.connpool != nil {
		c.connpool.Put(conn)
		return nil
	}
	return UnwrapConn(conn).Close()
}

// TCPSetKeepAlive change keep-alive when setsockopt after dial.
// WARN: Keep-alive is enable by default, ensure you know what you are doing
//       when you call this function and change it
func (c *Cluster) TCPSetKeepAlive(keepAlive bool) {
	c.tcpKeepAlive = keepAlive
}

// TCPSetKeepAlivePeriod change keep-alive period when setsockopt after dial.
// WARN: Keep-alive period is enable 3 seconds, ensure you know what you are doing
//       when you call this function and change it
func (c *Cluster) TCPSetKeepAlivePeriod(d time.Duration) {
	c.tcpKeepAlivePeriod = d
}

// TCPSetLinger change linger when setsockopt after dial
// WARN: Linger is enable by default, ensure you know what you are doing
//       when you call this function and change it
func (c *Cluster) TCPSetLinger(linger int) {
	c.tcpLinger = linger
}

// TCPSetNoDelay set NoDelay when setsockopt after dial
// WARN: NoDelay is enabled by default, ensure you know what you are doing
//       when you call this function and change it
func (c *Cluster) TCPSetNoDelay(noDelay bool) {
	c.tcpNoDelay = noDelay
}

// tcpsetsockopt
func (c *Cluster) tcpsetsockopt(conn *net.TCPConn) error {
	var err error

	// Keep-Alive
	err = conn.SetKeepAlive(c.tcpKeepAlive)
	if err != nil {
		return err
	}

	// Keep-Alive Period
	err = conn.SetKeepAlivePeriod(c.tcpKeepAlivePeriod)
	if err != nil {
		return err
	}

	// Linger
	err = conn.SetLinger(c.tcpLinger)
	if err != nil {
		return err
	}

	// NoDelay
	err = conn.SetNoDelay(c.tcpNoDelay)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cluster) Metrics() map[string]int64 {
	return map[string]int64{
		"dial_direct":     atomic.LoadInt64(&c.metricDialDirect),
		"dial_pool_reuse": atomic.LoadInt64(&c.metricDialPoolReuse),
	}
}
