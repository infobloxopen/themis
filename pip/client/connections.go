package client

import (
	"net"
	"sync"
	"sync/atomic"
)

const (
	pipClientIdle uint32 = iota
	pipClientConnecting
	pipClientConnected
	pipClientClosing
)

func (c *client) Connect() error {
	if !atomic.CompareAndSwapUint32(c.state, pipClientIdle, pipClientConnecting) {
		return ErrConnected
	}

	state := pipClientIdle
	defer func() {
		atomic.StoreUint32(c.state, state)
	}()

	n, err := net.Dial(c.opts.net, c.opts.addr)
	if err != nil {
		return err
	}

	conn := c.newConnection(n)
	conn.start()

	c.Lock()
	c.c = conn
	c.Unlock()

	state = pipClientConnected

	return nil
}

func (c *client) Close() {
	if !atomic.CompareAndSwapUint32(c.state, pipClientConnected, pipClientClosing) {
		return
	}
	defer atomic.StoreUint32(c.state, pipClientIdle)

	c.Lock()
	conn := c.c
	c.c = nil
	c.Unlock()

	conn.close()
}

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

func (c *connection) close() {
	c.n.Close()
	close(c.t)
	c.g.Wait()

	close(c.r)
	c.w.Wait()
}
