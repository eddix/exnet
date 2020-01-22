package exnet_test

import (
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eddix/exnet"
)

func TestConnPool(t *testing.T) {
	t.Run("Sync Conn Pool", func(t *testing.T) {
		testConnPool(t, false)
	})
	t.Run("Async Conn Pool", func(t *testing.T) {
		testConnPool(t, true)
	})
}

func testConnPool(t *testing.T, async bool) {
	cap, connsNum := 7, 100
	conf := &exnet.ConnPoolConfig{
		Cap: cap,
	}
	var pool exnet.ConnPool
	if async {
		pool = exnet.NewAsyncConnPool(conf)
	} else {
		pool = exnet.NewSyncConnPool(conf)
	}
	for i := 0; i < connsNum; i++ {
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
		var poolStat []string
		// static inspect pool for SyncConnPool
		if sp, ok := pool.(*exnet.SyncConnPool); ok {
			sp.Map(func(conn net.Conn) {
				if conn != nil {
					poolStat = append(poolStat, conn.LocalAddr().String())
				}
			})
			t.Logf("PoolStat-%d: %s", i, strings.Join(poolStat, ", "))
		}
	}
	// get
	var conn net.Conn
	var expectAddr = &net.TCPAddr{
		IP:   net.IP{0, 0, 0, 0},
		Port: 0,
		Zone: "",
	}
	for i := connsNum - cap; i < connsNum; i++ {
		expectAddr.Port = i
		conn = pool.Get()
		assert.Equal(t, expectAddr, conn.LocalAddr())
	}
	conn = pool.Get()
	assert.Nil(t, conn)
}

func BenchmarkConnPool(b *testing.B) {
	b.Run("SyncPutGet", func(b *testing.B) {
		benchmarkConnPool(b, false)
	})
	b.Run("AsyncPutGet", func(b *testing.B) {
		benchmarkConnPool(b, true)
	})
}

func benchmarkConnPool(b *testing.B, async bool) {
	conf := &exnet.ConnPoolConfig{
		Cap: b.N/2 + 1,
	}
	var pool exnet.ConnPool
	if async {
		pool = exnet.NewAsyncConnPool(conf)
	} else {
		pool = exnet.NewSyncConnPool(conf)
	}
	var wg sync.WaitGroup
	var got int64
	for i := 0; i < b.N; i++ {
		wg.Add(2)
		go func(port int) {
			conn := exnet.WithConn(&testConn{
				localAddr: func() net.Addr {
					return &net.TCPAddr{
						IP:   net.IP{0, 0, 0, 0},
						Port: port,
					}
				},
			})
			pool.Put(conn)
			wg.Done()
		}(i)
		go func() {
			conn := pool.Get()
			if conn != nil {
				atomic.AddInt64(&got, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	b.Logf("%s got/N = %d/%d\n", b.Name(), got, b.N)
}
