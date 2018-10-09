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

type simpleBalancer struct {
	sync.RWMutex

	c *connection
}

func (b *simpleBalancer) start(c *client) error {
	n, err := net.Dial(c.opts.net, c.opts.addr)
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
