package client

import (
	"net"
	"os"
	"sync"
)

type connection struct {
	i uint64

	c *client
	n net.Conn

	g sync.WaitGroup
	w sync.WaitGroup

	r chan request
	t chan struct{}
	p pipes
}

type destConn struct {
	d string
	c *connection
}

func (c *client) newConnection(n net.Conn) *connection {
	i := c.nextID()
	if i == 0 {
		return nil
	}

	return &connection{
		i: i,
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

func (c *connection) isFull() bool {
	return len(c.r) >= cap(c.r)
}

const (
	netConnRefusedMsg = "connection refused"
	netConnClosedMsg  = "use of closed network connection"
)

func isConnRefused(err error) bool {
	if err, ok := err.(*net.OpError); ok {
		if err, ok := err.Err.(*os.SyscallError); ok {
			return err.Err.Error() == netConnRefusedMsg
		}

		return false
	}

	return false
}

func isConnClosed(err error) bool {
	if err, ok := err.(*net.OpError); ok {
		return err.Err.Error() == netConnClosedMsg
	}

	return false
}

func isConnTimeout(err error) bool {
	if err, ok := err.(*net.OpError); ok {
		return err.Timeout()
	}

	return false
}
