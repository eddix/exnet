package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/eddix/exnet"
)

func main() {
	l, err := exnet.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	l.SetAcceptCallback(func(conn net.Conn) error {
		_ = exnet.TraceConn(conn, os.Stderr, nil)
		_ = conn.SetDeadline(time.Now().Add(time.Second))
		_ = conn.SetReadDeadline(time.Now().Add(time.Second))
		_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
		_ = exnet.Freeze(conn)
		return nil
	})
	panic(http.Serve(l, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
	})))
}
