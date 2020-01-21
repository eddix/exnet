package exnet

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

// ConnTracer handlers to trace exnet.Conn
type ConnTracer struct {
	Writer io.Writer

	TraceReadFunc             func(conn net.Conn, data []byte, err error)
	TraceWriteFunc            func(conn net.Conn, data []byte, err error)
	TraceCloseFunc            func(conn net.Conn, err error)
	TraceSetDeadlineFunc      func(conn net.Conn, t time.Time, err error)
	TraceSetReadDeadlineFunc  func(conn net.Conn, t time.Time, err error)
	TraceSetWriteDeadlineFunc func(conn net.Conn, t time.Time, err error)
}

// ReadTracer interface
type ReadTracer interface {
	TraceRead(conn net.Conn, data []byte, err error)
}

// TraceRead is function implement ReadTracer interface
type TraceRead func(conn net.Conn, data []byte, err error)

// TraceRead implement ReadTracer interface
func (f TraceRead) TraceRead(conn net.Conn, data []byte, err error) {
	f(conn, data, err)
}

// WriteTracer interface
type WriteTracer interface {
	TraceWrite(conn net.Conn, data []byte, err error)
}

// TraceWrite is function implement WriteTracer interface
type TraceWrite func(conn net.Conn, data []byte, err error)

// TraceWrite implement WriteTracer interface
func (f TraceWrite) TraceWrite(conn net.Conn, data []byte, err error) {
	f(conn, data, err)
}

// CloseTracer interface
type CloseTracer interface {
	TraceClose(conn net.Conn, err error)
}

// TraceClose is function implement CloseTracer interface
type TraceClose func(conn net.Conn, err error)

// TraceClose implement CloseTracer interface
func (f TraceClose) TraceClose(conn net.Conn, err error) {
	f(conn, err)
}

// SetDeadlineTracer interface
type SetDeadlineTracer interface {
	TraceSetDeadline(conn net.Conn, t time.Time, err error)
}

// TraceSetDeadline is function implement SetDeadlineTracer interface
type TraceSetDeadline func(conn net.Conn, t time.Time, err error)

// TraceSetDeadline implement SetDeadlineTracer interface
func (f TraceSetDeadline) TraceSetDeadline(conn net.Conn, t time.Time, err error) {
	f(conn, t, err)
}

// SetReadDeadlineTracer interface
type SetReadDeadlineTracer interface {
	TraceSetReadDeadline(conn net.Conn, t time.Time, err error)
}

// TraceSetReadDeadline is function implement SetReadDeadlineTracer interface
type TraceSetReadDeadline func(conn net.Conn, t time.Time, err error)

// TraceSetReadDeadline implement SetReadDeadlineTracer interface
func (f TraceSetReadDeadline) TraceSetReadDeadline(conn net.Conn, t time.Time, err error) {
	f(conn, t, err)
}

// SetWriteDeadlineTracer interface
type SetWriteDeadlineTracer interface {
	TraceSetWriteDeadline(conn net.Conn, t time.Time, err error)
}

// TraceSetWriteDeadline is function implement SetWriteDeadlineTracer interface
type TraceSetWriteDeadline func(conn net.Conn, t time.Time, err error)

// TraceSetWriteDeadline implement SetWriteDeadlineTracer interface
func (f TraceSetWriteDeadline) TraceSetWriteDeadline(conn net.Conn, t time.Time, err error) {
	f(conn, t, err)
}

var (
	_ ReadTracer             = &ConnTracer{}
	_ WriteTracer            = &ConnTracer{}
	_ CloseTracer            = &ConnTracer{}
	_ SetDeadlineTracer      = &ConnTracer{}
	_ SetReadDeadlineTracer  = &ConnTracer{}
	_ SetWriteDeadlineTracer = &ConnTracer{}
)

func (ct *ConnTracer) TraceRead(conn net.Conn, data []byte, err error) {
	if ct.TraceReadFunc != nil {
		ct.TraceReadFunc(conn, data, err)
		return
	}
	if err != nil {
		ct.logger().Printf("%s read error: %s", ct.connString(conn),
			err.Error())
	} else {
		if len(data) <= 32 {
			ct.logger().Printf("%s read %d bytes, HEX=%x",
				ct.connString(conn), len(data), data)
		} else {
			ct.logger().Printf("%s read %d bytes, HEX(First 32)=%x",
				ct.connString(conn), len(data), data[:32])
		}
	}
}

func (ct *ConnTracer) TraceWrite(conn net.Conn, data []byte, err error) {
	if ct.TraceWriteFunc != nil {
		ct.TraceWriteFunc(conn, data, err)
		return
	}
	logger := ct.logger()
	if err != nil {
		logger.Printf("%s write error: %s", ct.connString(conn),
			err.Error())
	} else {
		if len(data) <= 32 {
			logger.Printf("%s write %d bytes, HEX=%x",
				ct.connString(conn), len(data), data)
		} else {
			logger.Printf("%s write %d bytes, HEX(First 32)=%x",
				ct.connString(conn), len(data), data[:32])
		}
	}
}

func (ct *ConnTracer) TraceClose(conn net.Conn, err error) {
	if ct.TraceCloseFunc != nil {
		ct.TraceCloseFunc(conn, err)
		return
	}
	if err != nil {
		ct.logger().Printf("%s close error: %s", ct.connString(conn), err.Error())
	} else {
		ct.logger().Printf("%s close done", ct.connString(conn))
	}
}

func (ct *ConnTracer) TraceSetDeadline(conn net.Conn, t time.Time, err error) {
	if ct.TraceSetDeadlineFunc != nil {
		ct.TraceSetDeadlineFunc(conn, t, err)
		return
	}
	if err != nil {
		ct.logger().Printf("%s setDeadline %s error: %s", ct.connString(conn), t.String(), err.Error())
	} else {
		ct.logger().Printf("%s setDeadline %s done", ct.connString(conn), t.String())
	}
}

func (ct *ConnTracer) TraceSetReadDeadline(conn net.Conn, t time.Time, err error) {
	if ct.TraceSetReadDeadlineFunc != nil {
		ct.TraceSetReadDeadlineFunc(conn, t, err)
		return
	}
	if err != nil {
		ct.logger().Printf("%s setReadDeadline %s error: %s", ct.connString(conn), t.String(), err.Error())
	} else {
		ct.logger().Printf("%s setReadDeadline %s done", ct.connString(conn), t.String())
	}
}

func (ct *ConnTracer) TraceSetWriteDeadline(conn net.Conn, t time.Time, err error) {
	if ct.TraceSetWriteDeadlineFunc != nil {
		ct.TraceSetWriteDeadlineFunc(conn, t, err)
		return
	}
	if err != nil {
		ct.logger().Printf("%s setWriteDeadline %s error: %s", ct.connString(conn), t.String(), err.Error())
	} else {
		ct.logger().Printf("%s setWriteDeadline %s done", ct.connString(conn), t.String())
	}
}

func (ct *ConnTracer) connString(conn net.Conn) string {
	return fmt.Sprintf("(%s|%s)", conn.LocalAddr().String(), conn.RemoteAddr().String())
}

func (ct *ConnTracer) logger() *log.Logger {
	if ct.Writer == nil {
		return log.New(os.Stdout, "[exnet.Conn] ", log.Ldate|log.Ltime|log.Lmicroseconds)
	}
	return log.New(ct.Writer, "[exnet.Conn] ", log.Ldate|log.Ltime|log.Lmicroseconds)
}

// DebugConnTracer is a ConnTracer for common use
func DebugConnTracer(c net.Conn, w io.Writer) *ConnTracer {
	return &ConnTracer{}
}

// TraceConn add a tracer, if tracer is nil, use ConnTracer as default
func TraceConn(conn net.Conn, tracer interface{}) error {
	c, ok := conn.(*Conn)
	if !ok {
		return ErrNotExnetConn
	}
	if tracer == nil {
		tracer = &ConnTracer{}
	}
	c.tracer = tracer

	return nil
}

// StopTraceConn remove tracer
func StopTraceConn(conn net.Conn) error {
	c, ok := conn.(*Conn)
	if !ok {
		return ErrNotExnetConn
	}
	c.tracer = nil

	return nil
}
