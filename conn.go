package exnet

import (
	"net"
	"time"
)

// Conn ExNet Connection implements net.Conn
type Conn struct {
	// underlying net.Conn
	_conn net.Conn

	// timeout and deadline for read/write
	_readTimeout   time.Duration
	_readDeadline  time.Time
	_writeTimeout  time.Duration
	_writeDeadline time.Time
	_freeze        bool

	// trace handlers
	_traceRead             func(n int, err error, b []byte)
	_traceWrite            func(n int, err error, b []byte)
	_traceClose            func(err error)
	_traceSetDeadline      func(t time.Time, err error)
	_traceSetReadDeadline  func(t time.Time, err error)
	_traceSetWriteDeadline func(t time.Time, err error)
}

// Conn is an implementation of interface net.Conn
var _ net.Conn = &Conn{}

// Underlying net.Conn
func (c *Conn) Underlying() net.Conn {
	return c._conn
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c *Conn) Read(b []byte) (n int, err error) {
	n, err = c._conn.Read(b)
	// trace
	if c._traceRead != nil {
		c._traceRead(n, err, b)
	}

	return n, err
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c *Conn) Write(b []byte) (n int, err error) {
	n, err = c._conn.Write(b)
	// trace
	if c._traceWrite != nil {
		c._traceWrite(n, err, b)
	}

	return n, err
}

// Close the connection, ExNet will determine whether close by rules following:
// 1. if the service if configured with "Short Connection", close it.
// 2. if the service if configured with "Long Connection", will put it back to
//    preparation pool under situation following, otherwise close it.
//    a. The connection is not closed.
//    b. The number of idle connections doesn't reach limit.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c *Conn) Close() error {
	err := c._conn.Close()
	// trace
	if c._traceClose != nil {
		c._traceClose(err)
	}
	return err
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c._conn.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c._conn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
//
// Note that if a TCP connection has keep-alive turned on,
// which is the default unless overridden by Dialer.KeepAlive
// or ListenConfig.KeepAlive, then a keep-alive failure may
// also return a timeout error. On Unix systems a keep-alive
// failure on I/O can be detected using
// errors.Is(err, syscall.ETIMEDOUT).
func (c *Conn) SetDeadline(t time.Time) error {
	if c._freeze {
		return nil
	}
	err := c._conn.SetDeadline(t)
	if c._traceSetDeadline != nil {
		c._traceSetDeadline(t, err)
	}
	return err
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c *Conn) SetReadDeadline(t time.Time) error {
	if c._freeze {
		return nil
	}
	err := c._conn.SetReadDeadline(t)
	if c._traceSetReadDeadline != nil {
		c._traceSetReadDeadline(t, err)
	}
	return err
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	if c._freeze {
		return nil
	}
	err := c._conn.SetWriteDeadline(t)
	if c._traceSetWriteDeadline != nil {
		c._traceSetWriteDeadline(t, err)
	}
	return err
}

// WithConn wrap a net.Conn to exnet.Conn, return its-self if already wrapped
func WithConn(conn net.Conn) net.Conn {
	if _, ok := conn.(*Conn); ok {
		return conn
	}
	return &Conn{_conn: conn}
}

// Freeze changes of deadlines, call SetDeadline, SetReadDeadline, or SetWriteDeadline
// will do nothing on a freezed exnet.Conn, unless use Unfreeze() on the exnet.Conn
func Freeze(conn net.Conn) error {
	return changeFreeze(conn, true)
}

// Unfreeze an exnet.Conn which is freezed by Freeze()
func Unfreeze(conn net.Conn) error {
	return changeFreeze(conn, false)
}

func changeFreeze(conn net.Conn, freeze bool) error {
	if c, ok := conn.(*Conn); ok {
		c._freeze = freeze
		return nil
	}
	return ErrNotExnetConn
}