package exnet_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eddix/exnet"
)

func TestConnPool(t *testing.T) {
	size := 3
	pool := exnet.NewConnPool(&exnet.ConnPoolConfig{
		Size:  3,
		Async: false,
	})
	for i := 0; i < 10; i++ {
		port := i
		conn := exnet.WithConn(&testConn{
			localAddr: func() net.Addr {
				return &net.TCPAddr{
					IP:   net.IP{0, 0, 0, 0},
					Port: port,
					Zone: "",
				}
			},
		})
		pool.Put(conn)
	}
	// Should got 7, 8, 9
	var conn net.Conn
	var expectAddr = &net.TCPAddr{
		IP:   net.IP{0, 0, 0, 0},
		Port: 0,
		Zone: "",
	}
	for i := 10 - size; i < 10; i++ {
		expectAddr.Port = i
		conn = pool.Get()
		assert.Equal(t, expectAddr, conn.LocalAddr())
	}
	conn = pool.Get()
	assert.Nil(t, conn)
}
