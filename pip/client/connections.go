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

	nc, err := net.Dial(c.opts.net, c.opts.addr)
	if err != nil {
		return err
	}

	gwg := new(sync.WaitGroup)
	req := make(chan request, c.opts.maxQueue)
	wwg := new(sync.WaitGroup)

	p := makePipes(c.opts.maxQueue)

	wwg.Add(2)
	go c.writer(wwg, nc, req, p)
	go c.reader(wwg, nc, p)

	c.lock.Lock()
	c.c = nc
	c.gwg = gwg
	c.req = req
	c.wwg = wwg
	c.pipes = p
	c.lock.Unlock()

	state = pipClientConnected

	return nil
}

func (c *client) Close() {
	if !atomic.CompareAndSwapUint32(c.state, pipClientConnected, pipClientClosing) {
		return
	}
	defer atomic.StoreUint32(c.state, pipClientIdle)

	c.lock.Lock()
	nc := c.c
	c.c = nil
	gwg := c.gwg
	c.gwg = nil
	req := c.req
	c.req = nil
	wwg := c.wwg
	c.wwg = nil
	c.lock.Unlock()

	nc.Close()

	gwg.Wait()
	close(req)

	wwg.Wait()
}
