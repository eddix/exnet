package main

import (
	"context"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/eddix/exnet"
	"github.com/eddix/exnet/addresspicker"
)

func main() {
	ap := addresspicker.NewRoundRobin(nil)
	if err := ap.AppendTCPAddress("tcp", "www.baidu.com:443"); err != nil {
		log.Fatal(err)
	}
	cluster := exnet.Cluster{
		DialTimeout:   100 * time.Millisecond,
		ReadTimeout:   100 * time.Millisecond,
		WriteTimeout:  100 * time.Millisecond,
		AddressPicker: ap,
	}

	clusterClient := http.Client{
		Transport: &http.Transport{
			// Use ExNet.DialContext(serviceName string) function generator
			// DialContext: cluster.DialContext,

			// Or dial by hand to add
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				// network and addr is useless here
				log.Printf("Context: %v, network: %v, addr: %v\n", ctx, network, addr)
				// Real Dial happens here
				c, err := cluster.DialContext(ctx, network, addr)
				if err == nil {
					exnet.TraceConn(c, os.Stdout, nil)
				}
				return c, err
			},
		},
	}
	req, _ := http.NewRequest("GET", "https://www.baidu.com/", nil)
	resp, err := clusterClient.Do(req)
	if err == nil {
		b, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(b))
		resp.Body.Close()
	} else {
		log.Printf("Get failed: %s\n", err.Error())
	}
}
