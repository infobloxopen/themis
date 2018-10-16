package client

import (
	"sync"
	"sync/atomic"
)

type getter func(*uint64, []*connection) *connection

type provider struct {
	wg *sync.WaitGroup
	sync.RWMutex
	rCnd *sync.Cond
	wCnd *sync.Cond

	c *client

	started bool

	idx     *uint64
	queue   []*connection
	healthy map[string]*connection
	broken  map[uint64]destConn

	getter getter
}

func (p *provider) start(c *client, addrs []string) {
	p.Lock()
	defer p.Unlock()

	if p.started {
		return
	}
	p.started = true

	p.c = c

	p.getter = selectGetter(p.c.opts.balancer)

	p.idx = new(uint64)
	p.queue = []*connection{}
	p.healthy = make(map[string]*connection)
	p.broken = make(map[uint64]destConn)
	for _, a := range addrs {
		i := c.nextID()
		p.broken[i] = destConn{
			d: a,
		}
	}

	p.rCnd = sync.NewCond(p.RLocker())
	p.wCnd = sync.NewCond(p)

	p.wg = new(sync.WaitGroup)
	p.wg.Add(1)
	go p.connector(p.wg, p.c)
}

func (p *provider) stop() {
	p.Lock()

	if !p.started {
		p.Unlock()
		return
	}
	p.started = false

	p.queue = nil

	healthy := p.healthy
	p.healthy = nil

	broken := p.broken
	p.broken = nil

	wg := p.wg
	p.wg = nil

	rCnd := p.rCnd
	p.rCnd = nil

	wCnd := p.wCnd
	p.wCnd = nil

	p.getter = nil
	p.idx = nil
	p.c = nil

	p.Unlock()

	rCnd.Broadcast()
	wCnd.Signal()

	for _, c := range healthy {
		wg.Add(1)
		go func(c *connection) {
			defer wg.Done()

			c.close()
		}(c)
	}

	for _, dc := range broken {
		if dc.c != nil {
			wg.Add(1)
			go func(c *connection) {
				defer wg.Done()

				c.close()
			}(dc.c)
		}
	}

	wg.Wait()
}

func (p *provider) get() *connection {
	p.RLock()
	defer p.RUnlock()

	for p.started && len(p.queue) <= 0 {
		p.rCnd.Wait()
	}

	if p.started {
		if len(p.queue) > 0 {
			if c := p.getter(p.idx, p.queue); c != nil {
				c.g.Add(1)
				return c
			}
		}
	}

	return nil
}

func (p *provider) connector(wg *sync.WaitGroup, c *client) {
	defer wg.Done()

	for {
		a, conn := p.getBroken()
		if len(a) <= 0 && conn == nil {
			break
		}

		if conn != nil {
			conn.close()
		}

		if len(a) > 0 {
			wg.Add(1)
			go p.connect(wg, c, a)
		}
	}
}

func (p *provider) connect(wg *sync.WaitGroup, c *client, a string) {
	defer wg.Done()

	if n := c.dial(a); n != nil {
		p.Lock()
		defer p.Unlock()

		if !p.started {
			n.Close()
			return
		}

		conn := p.c.newConnection(n)
		conn.start()
		p.healthy[a] = conn
		p.queue = append(p.queue, conn)
		p.rCnd.Broadcast()
	}
}

func (p *provider) getBroken() (string, *connection) {
	p.Lock()
	defer p.Unlock()

	for p.started && len(p.broken) <= 0 {
		p.wCnd.Wait()
	}

	if p.started {
		for i, dc := range p.broken {
			delete(p.broken, i)
			return dc.d, dc.c
		}
	}

	return "", nil
}

func selectGetter(balancer int) getter {
	switch balancer {
	case balancerTypeRoundRobin:
		return roundRobinGetter

	case balancerTypeHotSpot:
		return hotSpotGetter
	}

	return simpleGetter
}

func simpleGetter(idx *uint64, queue []*connection) *connection {
	return queue[0]
}

func roundRobinGetter(idx *uint64, queue []*connection) *connection {
	return queue[int((atomic.AddUint64(idx, 1)-1)%uint64(len(queue)))]
}

func hotSpotGetter(idx *uint64, queue []*connection) *connection {
	total := uint64(len(queue))
	i := int(atomic.LoadUint64(idx) % total)
	for j := 0; j < len(queue); j++ {
		if c := queue[i]; c != nil && !c.isFull() {
			return c
		}

		i = int(atomic.AddUint64(idx, 1) % total)
	}

	return queue[i]
}
