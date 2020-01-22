package exnet

import (
	"net"
	"sync"
)

// ConnPool interface definition for connpool
type ConnPool interface {
	Put(net.Conn)
	Get() net.Conn
}

// ConnPoolConfig config for ConnPool
type ConnPoolConfig struct {
	Size int
}

// SyncConnPool maintain connections in pool,
// Get and Put manipulation is blocking until done.
type SyncConnPool struct {
	pool  []net.Conn
	ridx  int
	widx  int
	rwmtx sync.Mutex
	size  int
}

// connWrapper meanings the conn has underlying conn
type connWrapper interface {
	Underlying() net.Conn
}

func unwrapConn(conn net.Conn) net.Conn {
	uc, ok := conn, true
	for ok {
		if _, ok = uc.(connWrapper); ok {
			uc = uc.(connWrapper).Underlying()
		}
	}
	return uc
}

// NewSyncConnPool create a sync conn pool
func NewSyncConnPool(conf *ConnPoolConfig) ConnPool {
	if conf == nil {
		panic("ConnPoolConfig can't be nil")
	}
	p := &SyncConnPool{
		pool: make([]net.Conn, conf.Size),
		size: conf.Size,
	}
	return p
}

func (p *SyncConnPool) Put(conn net.Conn) {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	old := p.pool[p.widx]
	p.pool[p.widx] = unwrapConn(conn)
	if old != nil {
		_ = old.Close()
		p.ridx = p.grow(p.ridx)
	}
	p.widx = p.grow(p.widx)
}

func (p *SyncConnPool) Get() net.Conn {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	if p.pool[p.ridx] == nil {
		return nil
	}
	conn := p.pool[p.ridx]
	p.pool[p.ridx] = nil
	p.ridx = p.grow(p.ridx)
	return conn
}

func (p *SyncConnPool) Map(f func(net.Conn)) {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	for _, c := range p.pool {
		f(c)
	}
}

func (p *SyncConnPool) grow(n int) int {
	if n+1 < p.size {
		return n + 1
	}
	return 0
}

// AsyncConnPool use channel as connection pool
// Get and Put manipulations are channel in and out
type AsyncConnPool struct {
	pool chan net.Conn
}

func NewAsyncConnPool(conf *ConnPoolConfig) ConnPool {
	p := &AsyncConnPool{
		pool: make(chan net.Conn, conf.Size),
	}
	return p
}

func (p *AsyncConnPool) Get() net.Conn {
	select {
	case conn := <-p.pool:
		return conn
	default:
		return nil
	}
}

func (p *AsyncConnPool) Put(conn net.Conn) {
	for {
		done := false
		select {
		case p.pool <- conn:
			done = true
		default:
		}
		if done {
			break
		}
		<-p.pool
	}
}
