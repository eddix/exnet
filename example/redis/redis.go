package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/go-redis/redis"

	"github.com/eddix/exnet"
	"github.com/eddix/exnet/addresspicker"
)

func main() {
	ap := addresspicker.NewRoundRobin(nil)
	_ = ap.AppendTCPAddress("tcp", "localhost:6377") // wrong address
	_ = ap.AppendTCPAddress("tcp", "localhost:6378") // wrong address
	_ = ap.AppendTCPAddress("tcp", "localhost:6379")
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
	client := redis.NewClient(&redis.Options{
		Dialer: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var conn net.Conn
			var err error
			for attempt := 0; attempt < 3; attempt++ {
				conn, err = cluster.DialContext(ctx, network, addr)
				if err == nil {
					_ = exnet.TraceConn(conn, tracer)
					break
				}
			}
			return conn, err
		},
		Password: "",
		DB:       0,
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)
}
