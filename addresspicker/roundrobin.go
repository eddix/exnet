package addresspicker

import (
	"net"
	"sync"
)

// RoundRobin load balance strategy
type RoundRobin struct {
	addrs []net.Addr
	idx   int
	mtx   sync.Mutex
}

// NewRoundRobin address picker
func NewRoundRobin(addrs []net.Addr) *RoundRobin {
	return &RoundRobin{
		addrs: addrs,
		idx:   -1,
	}
}

// AppendTCPAddress append tcp address
func (rr *RoundRobin) AppendTCPAddress(network, address string) error {
	addr, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return err
	}
	rr.appendAddr(addr)
	return nil
}

// Addr return a net address
func (rr *RoundRobin) Addr() net.Addr {
	rr.mtx.Lock()
	defer rr.mtx.Unlock()

	rr.idx++
	if rr.idx == len(rr.addrs) {
		rr.idx = 0
	}
	return rr.addrs[rr.idx]
}

func (rr *RoundRobin) appendAddr(addr net.Addr) {
	rr.mtx.Lock()
	defer rr.mtx.Unlock()

	rr.addrs = append(rr.addrs, addr)
}
