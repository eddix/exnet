package exnet

import (
	"net"
	"sync"
	"time"
)

// ConnPool interface definition for connpool
type ConnPool interface {
	Put(net.Conn)
	Get() net.Conn
	Cap() int
	Size() int
	CloseAll()
}

// ConnPoolConfig config for ConnPool
type ConnPoolConfig struct {
	// Cap is the max capacity of the pool
	Cap int
	// TODO:
	// IdleTimeout is the max duration a connection is valid after put into pool,
	// after idle duration the connection will be closed and drop out.
	// 0 is no timeout.
	IdleTimeout time.Duration
}

// SyncConnPool maintain connections in pool,
// Get and Put manipulation is blocking until done.
type SyncConnPool struct {
	pool  []net.Conn
	ridx  int
	widx  int
	rwmtx sync.Mutex
	cap   int
	size  int
	idle  time.Duration
}

// TODO: change conn to connInPool to track status of every connection
type connInPool struct {
	conn  net.Conn
	putAt time.Time
}

// NewSyncConnPool create a sync conn pool
func NewSyncConnPool(conf *ConnPoolConfig) ConnPool {
	if conf == nil {
		panic("ConnPoolConfig can't be nil")
	}
	return &SyncConnPool{
		pool: make([]net.Conn, conf.Cap),
		cap:  conf.Cap,
	}
}

// Put connection into pool, replace the oldest connection
// in the pool if pool is full.
func (p *SyncConnPool) Put(conn net.Conn) {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	old := p.pool[p.widx]
	p.pool[p.widx] = UnwrapConn(conn)
	if old != nil {
		_ = old.Close()
		p.ridx = p.grow(p.ridx)
		p.size--
	}
	p.widx = p.grow(p.widx)
	p.size++
}

// Get connection from pool, return nil if pool is empty.
func (p *SyncConnPool) Get() net.Conn {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	if p.pool[p.ridx] == nil {
		return nil
	}
	conn := p.pool[p.ridx]
	p.pool[p.ridx] = nil
	p.ridx = p.grow(p.ridx)
	p.size--
	return conn
}

// Map iterate every connection in pool with f.
func (p *SyncConnPool) Map(f func(net.Conn)) {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	for _, c := range p.pool {
		f(c)
	}
}

// Cap return capacity of the pool
func (p *SyncConnPool) Cap() int {
	return p.cap
}

// Size return current size of the pool
func (p *SyncConnPool) Size() int {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()
	return p.size
}

// CloseAll close all connection in pool
func (p *SyncConnPool) CloseAll() {
	for p.Size() > 0 {
		conn := p.Get()
		if conn != nil {
			_ = conn.Close()
		}
	}
}

func (p *SyncConnPool) grow(n int) int {
	if n+1 < p.cap {
		return n + 1
	}
	return 0
}

// AsyncConnPool use channel as connection pool
// Get and Put manipulations are channel in and out
type AsyncConnPool struct {
	pool chan net.Conn
}

// NewAsyncConnPool create new AsyncConnPool
func NewAsyncConnPool(conf *ConnPoolConfig) ConnPool {
	if conf == nil {
		panic("ConnPoolConfig can't be nil")
	}
	return &AsyncConnPool{
		pool: make(chan net.Conn, conf.Cap),
	}
}

// Get connection from pool, return nil if pool is empty.
func (p *AsyncConnPool) Get() net.Conn {
	select {
	case conn := <-p.pool:
		return conn
	default:
		return nil
	}
}

// Put connection into pool, replace the oldest connection
// in the pool if pool is full.
func (p *AsyncConnPool) Put(conn net.Conn) {
	for {
		done := false
		select {
		case p.pool <- UnwrapConn(conn):
			done = true
		default:
		}
		if done {
			break
		}
		<-p.pool
	}
}

func (p *AsyncConnPool) Cap() int {
	return cap(p.pool)
}

func (p *AsyncConnPool) Size() int {
	return len(p.pool)
}

// CloseAll close all connection in pool
func (p *AsyncConnPool) CloseAll() {
	for p.Size() > 0 {
		conn := p.Get()
		if conn != nil {
			_ = conn.Close()
		}
	}
}
