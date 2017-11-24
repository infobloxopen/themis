package pep

import (
	"sync"
	"sync/atomic"
)

const (
	crpIdle uint32 = iota
	crpWorking
	crpFull
	crpStopping
)

type connRetryPool struct {
	state *uint32

	ch    chan *streamConn
	count *uint32

	c *sync.Cond
	m *sync.RWMutex
}

func newConnRetryPool(conns []*streamConn) *connRetryPool {
	state := crpIdle

	ch := make(chan *streamConn, len(conns))
	for _, c := range conns {
		ch <- c
	}
	count := uint32(len(conns))

	return &connRetryPool{
		state: &state,
		ch:    ch,
		count: &count,
		c:     sync.NewCond(new(sync.Mutex)),
		m:     new(sync.RWMutex),
	}
}

func (p *connRetryPool) tryStart() {
	if atomic.CompareAndSwapUint32(p.state, crpIdle, crpFull) {
		go p.worker()
	}
}

func (p *connRetryPool) stop() {
	atomic.StoreUint32(p.state, crpStopping)
	p.c.Broadcast()

	p.m.Lock()
	defer p.m.Unlock()

	close(p.ch)
	p.ch = nil
}

func (p *connRetryPool) check() bool {
	return atomic.LoadUint32(p.state) == crpWorking
}

func (p *connRetryPool) wait() bool {
	p.c.L.Lock()
	defer p.c.L.Unlock()

	for {
		switch atomic.LoadUint32(p.state) {
		case crpStopping:
			return false

		case crpWorking:
			return true
		}

		p.c.Wait()
	}
}

func (p *connRetryPool) put(c *streamConn) {
	if c.markDisconnected() {
		p.m.RLock()
		defer p.m.RUnlock()

		if p.ch == nil {
			return
		}

		if atomic.AddUint32(p.count, 1) == uint32(cap(p.ch)) {
			atomic.CompareAndSwapUint32(p.state, crpWorking, crpFull)
		}

		p.ch <- c
	}
}

func (p *connRetryPool) worker() {
	for c := range p.ch {
		go p.reconnect(c)
	}
}

func (p *connRetryPool) reconnect(c *streamConn) {
	if err := c.connect(); err != nil {
		return
	}

	atomic.AddUint32(p.count, ^uint32(0))
	if atomic.CompareAndSwapUint32(p.state, crpFull, crpWorking) {
		p.c.Broadcast()
	}
}
