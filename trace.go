package exnet

// TODO: trace应该参考golang.org/x/net/trace的接口设计，尽可能保持一致。

import (
	"io"
	"log"
	"net"
	"time"
)

// ConnTracer handlers to trace exnet.Conn
type ConnTracer struct {
	TraceRead             func(n int, err error, b []byte)
	TraceWrite            func(n int, err error, b []byte)
	TraceClose            func(err error)
	TraceSetDeadline      func(t time.Time, err error)
	TraceSetReadDeadline  func(t time.Time, err error)
	TraceSetWriteDeadline func(t time.Time, err error)
}

// DebugConnTracer is a ConnTracer for common use
func DebugConnTracer(c net.Conn, w io.Writer) *ConnTracer {
	logger := log.New(w, "[exnet.Conn] ", log.Ldate|log.Ltime|log.Lmicroseconds)

	return &ConnTracer{
		TraceRead: func(n int, err error, b []byte) {
			if err != nil {
				logger.Printf("Read from %s error: %s\n", c.RemoteAddr().String(),
					err.Error())
			} else {
				if n <= 32 {
					logger.Printf("Read %d bytes from %s, HEX=%x\n",
						n, c.RemoteAddr().String(), b[:n])
				} else {
					logger.Printf("Read %d bytes from %s, HEX(First 32)=%x\n",
						n, c.RemoteAddr().String(), b[:32])
				}
			}
		},
		TraceWrite: func(n int, err error, b []byte) {
			if err != nil {
				logger.Printf("Write to %s error: %s\n", c.RemoteAddr().String(),
					err.Error())
			} else {
				if n <= 32 {
					logger.Printf("Write %d bytes to %s, HEX=%x\n",
						n, c.RemoteAddr().String(), b)
				} else {
					logger.Printf("Write %d bytes to %s, HEX(First 32)=%x\n",
						n, c.RemoteAddr().String(), b[:32])
				}
			}
		},
		TraceClose: func(err error) {
			if err != nil {
				logger.Printf("Close conn error: %s\n", err.Error())
			} else {
				logger.Printf("Close conn done")
			}
		},
		TraceSetDeadline: func(t time.Time, err error) {
			if err != nil {
				logger.Printf("SetDeadline %s error: %s\n", t.String(), err.Error())
			} else {
				logger.Printf("SetDeadline %s done\n", t.String())
			}
		},
		TraceSetReadDeadline: func(t time.Time, err error) {
			if err != nil {
				logger.Printf("SetReadDeadline %s error: %s\n", t.String(), err.Error())
			} else {
				logger.Printf("SetReadDeadline %s done\n", t.String())
			}
		},
		TraceSetWriteDeadline: func(t time.Time, err error) {
			if err != nil {
				logger.Printf("SetWriteDeadline %s error: %s\n", t.String(), err.Error())
			} else {
				logger.Printf("SetWriteDeadline %s done\n", t.String())
			}
		},
	}
}

// TraceConn add a tracer
func TraceConn(conn net.Conn, w io.Writer, ctfactory func(net.Conn, io.Writer) *ConnTracer) error {
	c, ok := conn.(*Conn)
	if !ok {
		return ErrNotExnetConn
	}
	var ct *ConnTracer
	if ctfactory == nil {
		ctfactory = DebugConnTracer
	}
	ct = ctfactory(conn, w)
	c._traceRead = ct.TraceRead
	c._traceWrite = ct.TraceWrite
	c._traceClose = ct.TraceClose
	c._traceSetDeadline = ct.TraceSetDeadline
	c._traceSetReadDeadline = ct.TraceSetReadDeadline
	c._traceSetWriteDeadline = ct.TraceSetWriteDeadline

	return nil
}

// StopTraceConn remove tracer
func StopTraceConn(conn net.Conn) error {
	c, ok := conn.(*Conn)
	if !ok {
		return ErrNotExnetConn
	}
	c._traceRead = nil
	c._traceWrite = nil
	c._traceClose = nil
	c._traceSetDeadline = nil
	c._traceSetReadDeadline = nil
	c._traceSetWriteDeadline = nil

	return nil
}
