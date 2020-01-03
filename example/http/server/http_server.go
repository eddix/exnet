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
		exnet.TraceConn(conn, os.Stderr, nil)
		conn.SetDeadline(time.Now().Add(time.Second))
		conn.SetReadDeadline(time.Now().Add(time.Second))
		conn.SetWriteDeadline(time.Now().Add(time.Second))
		exnet.Freeze(conn)
		return nil
	})
	var f http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "hello world")
	}
	http.Serve(l, f)
}
