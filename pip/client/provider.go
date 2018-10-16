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
	retry   map[string]chan struct{}

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
	p.retry = make(map[string]chan struct{})

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

	retry := p.retry
	p.retry = nil

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

	for _, ch := range retry {
		close(ch)
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
		a, conn, ch := p.getBroken()
		if len(a) <= 0 && conn == nil {
			break
		}

		if conn != nil {
			conn.close()
		}

		if len(a) > 0 {
			wg.Add(1)
			go p.connect(wg, c, a, ch)
		}
	}
}

func (p *provider) connect(wg *sync.WaitGroup, c *client, a string, ch <-chan struct{}) {
	defer wg.Done()

	n := c.dial(a, ch)
	p.Lock()
	defer p.Unlock()

	if !p.started {
		if n != nil {
			if err := n.Close(); err != nil && c.opts.onErr != nil {
				c.opts.onErr(n.RemoteAddr(), err)
			}
		}

		return
	}

	delete(p.retry, a)

	if n != nil {
		conn := p.c.newConnection(n)
		conn.start()

		p.healthy[a] = conn
		p.queue = append(p.queue, conn)

		p.rCnd.Broadcast()
	}
}

func (p *provider) getBroken() (string, *connection, <-chan struct{}) {
	p.Lock()
	defer p.Unlock()

	for p.started && len(p.broken) <= 0 {
		p.wCnd.Wait()
	}

	if p.started {
		for i, dc := range p.broken {
			delete(p.broken, i)

			if len(dc.d) > 0 {
				if _, ok := p.retry[dc.d]; !ok {
					ch := make(chan struct{})
					p.retry[dc.d] = ch
					return dc.d, dc.c, ch
				}
			}

			return "", dc.c, nil
		}
	}

	return "", nil, nil
}

func (p *provider) unqueue(c *connection) {
	for i, qc := range p.queue {
		if c == qc {
			if i == 0 {
				p.queue = p.queue[i+1:]
			} else if i == len(p.queue)-1 {
				p.queue = p.queue[:i]
			} else {
				p.queue = append(p.queue[:i], p.queue[i+1:]...)
			}

			return
		}
	}
}

func (p *provider) unhealthy(c *connection) string {
	for d, hc := range p.healthy {
		if c == hc {
			delete(p.healthy, d)
			return d
		}
	}

	return ""
}

func (p *provider) report(c *connection) {
	p.Lock()
	defer p.Unlock()

	if !p.started {
		return
	}

	if _, ok := p.broken[c.i]; ok {
		return
	}

	p.unqueue(c)
	d := p.unhealthy(c)

	p.broken[c.i] = destConn{
		d: d,
		c: c,
	}
	p.wCnd.Signal()
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
