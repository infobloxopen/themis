package client

import (
	"net"
	"sync"
)

type connection struct {
	c *client
	n net.Conn

	g sync.WaitGroup
	w sync.WaitGroup

	r chan request
	t chan struct{}
	p pipes
}

func (c *client) newConnection(n net.Conn) *connection {
	return &connection{
		c: c,
		n: n,
		r: make(chan request, c.opts.maxQueue),
		t: make(chan struct{}),
		p: makePipes(c.opts.maxQueue, c.opts.timeout.Nanoseconds()),
	}
}

func (c *connection) start() {
	c.w.Add(3)
	go c.writer()
	go c.reader()
	go c.terminator()
}

func (c *connection) closeNet() {
	if err := c.n.Close(); err != nil && !isConnClosed(err) {
		if f := c.c.opts.onErr; f != nil {
			f(c.n.RemoteAddr(), err)
		}
	}
}

func (c *connection) close() {
	c.closeNet()

	close(c.t)
	c.g.Wait()

	close(c.r)
	c.w.Wait()
}

func (c *connection) get(b []byte) ([]byte, error) {
	i, p := c.p.alloc()
	defer c.p.free(i)

	c.r <- request{
		i: i,
		b: b,
	}

	return p.get()
}

const netConnClosedMsg = "use of closed network connection"

func isConnClosed(err error) bool {
	switch err := err.(type) {
	case *net.OpError:
		return err.Err.Error() == netConnClosedMsg
	}

	return false
}
