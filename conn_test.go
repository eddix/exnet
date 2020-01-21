package exnet_test

import (
	"net"
	"time"
)

type testConn struct {
	read             func(b []byte) (n int, err error)
	write            func(b []byte) (n int, err error)
	close            func() error
	localAddr        func() net.Addr
	remoteAddr       func() net.Addr
	setDeadline      func(t time.Time) error
	setReadDeadline  func(t time.Time) error
	setWriteDeadline func(t time.Time) error
}

var _ net.Conn = &testConn{}

func (c *testConn) Read(b []byte) (int, error) {
	if c.read == nil {
		return 0, nil
	}
	return c.read(b)
}

func (c *testConn) Write(b []byte) (int, error) {
	if c.write == nil {
		return len(b), nil
	}
	return c.write(b)
}

func (c *testConn) Close() error {
	if c.close == nil {
		return nil
	}
	return c.close()
}

func (c *testConn) LocalAddr() net.Addr {
	if c.localAddr == nil {
		return &net.TCPAddr{
			IP:   net.IP{127, 0, 0, 1},
			Port: 0,
			Zone: "",
		}
	}
	return c.localAddr()
}

func (c *testConn) RemoteAddr() net.Addr {
	if c.remoteAddr == nil {
		return &net.TCPAddr{
			IP:   net.IP{127, 0, 0, 1},
			Port: 0,
			Zone: "",
		}
	}
	return c.remoteAddr()
}

func (c *testConn) SetDeadline(t time.Time) error {
	if c.setDeadline == nil {
		return nil
	}
	return c.setDeadline(t)
}

func (c *testConn) SetReadDeadline(t time.Time) error {
	if c.setReadDeadline == nil {
		return nil
	}
	return c.setReadDeadline(t)
}

func (c *testConn) SetWriteDeadline(t time.Time) error {
	if c.setWriteDeadline == nil {
		return nil
	}
	return c.setWriteDeadline(t)
}
