package exnet_test

import (
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eddix/exnet"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestListenAndDial(t *testing.T) {
	lis, err := exnet.Listen("tcp", ":0")
	if err != nil {
		t.Fatal("Can't listen ", err)
	}
	msg := "hello world"

	// server side
	lisch := make(chan struct{})
	go func() {
		defer func() {
			assert.NoError(t, lis.Close())
			close(lisch)
		}()
		lis.SetAcceptCallback(func(conn net.Conn) error {
			assert.NoError(t, exnet.TraceConn(conn, nil))
			return nil
		})
		conn, err := lis.Accept()
		assert.NoError(t, err)
		buf := make([]byte, len(msg))
		assert.NoError(t, conn.SetReadDeadline(time.Now().Add(10*time.Millisecond)))
		assert.NoError(t, exnet.Freeze(conn))
		n, err := io.ReadFull(conn, buf)
		assert.NoError(t, err)
		assert.NoError(t, exnet.Unfreeze(conn))
		assert.Equal(t, n, len(msg))
		assert.Equal(t, msg, string(buf))
		assert.NoError(t, conn.Close())
		assert.NoError(t, exnet.StopTraceConn(conn))
	}()

	// client side
	conn, err := exnet.Dial("tcp", lis.Addr().String())
	assert.NoError(t, err)
	assert.NoError(t, exnet.TraceConn(conn, nil))
	assert.NoError(t, conn.SetWriteDeadline(time.Now().Add(10*time.Millisecond)))
	assert.NoError(t, exnet.Freeze(conn))
	n, err := conn.Write([]byte(msg))
	assert.NoError(t, err)
	assert.NoError(t, exnet.Unfreeze(conn))
	assert.Equal(t, n, len(msg))
	assert.NoError(t, conn.Close())
	assert.NoError(t, exnet.StopTraceConn(conn))

	// wait for listen stop
	<-lisch
}
