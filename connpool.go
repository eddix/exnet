package exnet

import (
	"net"
	"sync"
)

// ConnPool maintain connections in pool
type ConnPool struct {
	pool  []net.Conn
	ridx  int
	widx  int
	rwmtx sync.Mutex
	size  int

	async bool

	inch  chan net.Conn
	outch chan net.Conn
	stop  chan struct{}
}

type ConnPoolConfig struct {
	Size  int
	Async bool
}

// connWrapper meanings the conn has underlying conn
type connWrapper interface {
	Underlying() net.Conn
}

func NewConnPool(conf *ConnPoolConfig) *ConnPool {
	if conf == nil {
		panic("ConnPoolConfig can't be nil")
	}
	p := &ConnPool{
		pool:  make([]net.Conn, conf.Size),
		size:  conf.Size,
		async: conf.Async,
	}
	if conf.Async {
		p.inch = make(chan net.Conn, conf.Size/2)
		p.outch = make(chan net.Conn, conf.Size/2)
		p.stop = make(chan struct{})
	}
	return p
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

func (p *ConnPool) Put(conn net.Conn) {
	if !p.async {
		p.in(conn)
		return
	}
	select {
	case p.inch <- conn:
	default: // ignore if p.inch is blocking
	}
}

func (p *ConnPool) Get() net.Conn {
	if !p.async {
		return p.out()
	}
	select {
	case ret := <-p.outch:
		return ret
	default:
	}
	return nil
}

func (p *ConnPool) Map(f func(net.Conn)) {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	for _, c := range p.pool {
		f(c)
	}
}

func (p *ConnPool) putLoop() {
	for {
		select {
		case <-p.stop:
			return
		case conn := <-p.inch:
			p.in(conn)
		}
	}
}

func (p *ConnPool) getLoop() {
	for {
		select {
		case <-p.stop:
			return
		default:
		}
		p.outch <- p.out()
	}
}

func (p *ConnPool) in(conn net.Conn) {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	old := p.pool[p.widx]
	p.pool[p.widx] = unwrapConn(conn)
	p.widx = p.grow(p.widx)
	if old != nil {
		p.ridx = p.grow(p.ridx)
		_ = old.Close()
	}
}

func (p *ConnPool) out() net.Conn {
	p.rwmtx.Lock()
	defer p.rwmtx.Unlock()

	if p.pool[p.ridx] == nil {
		return nil
	}
	conn := p.pool[p.ridx]
	p.pool[p.ridx] = nil
	p.ridx++
	if p.ridx == p.size {
		p.ridx = 0
	}
	return conn
}

func (p *ConnPool) grow(n int) int {
	n++
	if n < p.size {
		return n
	}
	return 0
}
