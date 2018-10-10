package client

import (
	"net"
	"sync"
)

type balancer interface {
	start(c *client) error
	stop()
	get() *connection
}

func newBalancer(network string, balancerType int) balancer {
	if balancerType == balancerTypeRoundRobin && network != unixNet {
		return new(roundRobinBalancer)
	}

	return new(simpleBalancer)
}

type simpleBalancer struct {
	sync.RWMutex

	c *connection
}

func (b *simpleBalancer) start(c *client) error {
	n, err := net.DialTimeout(c.opts.net, c.opts.addr, c.opts.connTimeout)
	if err != nil {
		return err
	}

	conn := c.newConnection(n)
	conn.start()

	b.Lock()
	b.c = conn
	b.Unlock()

	return nil
}

func (b *simpleBalancer) stop() {
	b.Lock()
	c := b.c
	b.c = nil
	b.Unlock()

	c.close()
}

func (b *simpleBalancer) get() *connection {
	b.RLock()
	defer b.RUnlock()

	c := b.c
	if c != nil {
		c.g.Add(1)
	}

	return c
}
