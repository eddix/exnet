package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis"

	"github.com/eddix/exnet"
	"github.com/eddix/exnet/addresspicker"
)

func main() {
	// Create an exnet cluster
	ap := addresspicker.NewRoundRobin(nil)
	_ = ap.AppendTCPAddress("tcp", "localhost:6377") // wrong address
	_ = ap.AppendTCPAddress("tcp", "localhost:6378") // wrong address
	_ = ap.AppendTCPAddress("tcp", "localhost:6379") // right address
	cluster := &exnet.Cluster{
		DialTimeout:   10 * time.Millisecond,
		ReadTimeout:   100 * time.Millisecond,
		WriteTimeout:  100 * time.Millisecond,
		AddressPicker: ap,
	}
	// use custom tracer
	tracer := &exnet.ConnTracer{}
	tracer.TraceReadFunc = func(conn net.Conn, data []byte, err error) {
		if err != nil {
			log.Printf("Read Redis Error: %s", err.Error())
			return
		}
		log.Printf("Read Redis: (%s)", string(data))
	}
	tracer.TraceWriteFunc = func(conn net.Conn, data []byte, err error) {
		if err != nil {
			log.Printf("Write Redis Error: %s", err.Error())
			return
		}
		log.Printf("Write Redis: (%s)", string(data))
	}
	// create client with custom Dialer
	rdb := redis.NewClient(&redis.Options{
		Dialer: func() (net.Conn, error) {
			var conn net.Conn
			var err error
			for attempt := 0; attempt < 3; attempt++ {
				conn, err = cluster.Dial("", "")
				if err == nil {
					_ = exnet.TraceConn(conn, tracer)
					break
				}
			}
			return conn, err
		},
	})
	pong, err := rdb.Ping().Result()
	fmt.Println(pong, err)
}
