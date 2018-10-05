package client

import (
	"net"
	"sync"
	"time"
)

func (c *client) terminator(wg *sync.WaitGroup, nc net.Conn, p pipes, done chan struct{}) {
	defer wg.Done()

	ch := c.opts.termFlushCh
	if ch == nil {
		t := time.NewTicker(c.opts.termInt)
		defer t.Stop()

		ch = t.C
	}

	for {
		select {
		case <-done:
			p.flush()
			return

		case t := <-ch:
			if p.check(t) {
				nc.Close()
			}
		}
	}
}
