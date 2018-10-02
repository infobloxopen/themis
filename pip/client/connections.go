package client

import (
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

	rwg := new(sync.WaitGroup)
	req := make(chan request, c.opts.maxQueue)
	wwg := new(sync.WaitGroup)

	p := makePipes(c.opts.maxQueue)

	wwg.Add(1)
	go c.writer(wwg, req, p)

	c.lock.Lock()
	c.rwg = rwg
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
	rwg := c.rwg
	c.rwg = nil
	req := c.req
	c.req = nil
	wwg := c.wwg
	c.wwg = nil
	c.lock.Unlock()

	rwg.Wait()
	close(req)

	wwg.Wait()
}
