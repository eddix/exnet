package exnet_test

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/eddix/exnet"
	"github.com/eddix/exnet/addresspicker"
)

var (
	smsg = []byte("hello server")
	cmsg = []byte("hello client")
)

type testServer struct {
	listener net.Listener
	stopch   chan struct{}
}

func (s *testServer) run(t *testing.T) {
	var err error
	s.listener, err = exnet.Listen("tcp", ":0")
	assert.Nil(t, err)
	go func() {
		for {
			select {
			case <-s.stopch:
				return
			default:
			}
			conn, err := s.listener.Accept()
			if err != nil {
				continue
			}
			go func() {
				// read test
				buf := make([]byte, len(cmsg))
				n, err := conn.Read(buf)
				assert.Equal(t, len(cmsg), n)
				assert.Equal(t, cmsg, buf)
				assert.NoError(t, err)
				// write test
				n, err = conn.Write(smsg)
				assert.Equal(t, len(smsg), n)
				assert.NoError(t, err)
				// close test
				assert.NoError(t, conn.Close())
			}()
		}
	}()
}

func (s *testServer) stop() {
	close(s.stopch)
}

func makeServers(t *testing.T, amount int) []*testServer {
	srvs := make([]*testServer, amount)
	for i := 0; i < amount; i++ {
		srv := &testServer{
			stopch: make(chan struct{}),
		}
		srv.run(t)
		srvs[i] = srv
	}
	return srvs
}

func TestCluster(t *testing.T) {
	t.Run("SimpleTest", testCluster)
}

func testCluster(t *testing.T) {
	srvs := makeServers(t, 100)
	cluster := exnet.NewCluster(&exnet.ClusterConfig{
		DialTimeout:  100 * time.Millisecond,
		ReadTimeout:  500 * time.Microsecond,
		WriteTimeout: 500 * time.Microsecond,
		PoolConfig: &exnet.ConnPoolConfig{
			Cap: 100,
		},
		UseAsyncPool: true,
	})
	ap := addresspicker.NewRoundRobin(nil)
	for _, s := range srvs {
		assert.NotNil(t, s.listener.Addr())
		ap.AppendTCPAddress(s.listener.Addr().Network(), s.listener.Addr().String())
	}
	cluster.AddressPicker = ap

	for i := 0; i < 1000; i++ {
		conn, err := cluster.Dial("", "")
		assert.NoError(t, err)
		assert.NotNil(t, conn)
		// write test
		n, err := conn.Write(cmsg)
		assert.Equal(t, len(cmsg), n)
		assert.NoError(t, err)
		// read test
		buf := make([]byte, len(smsg))
		n, err = conn.Read(buf)
		assert.Equal(t, smsg, buf)
		assert.NoError(t, err)
		// close test
		assert.NoError(t, conn.Close())
	}

	t.Logf("dial metrics: %v", cluster.Metrics())
}
