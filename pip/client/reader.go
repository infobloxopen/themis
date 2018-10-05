package client

import (
	"net"
	"sync"
)

func (c *client) reader(wg *sync.WaitGroup, nc net.Conn, p pipes) {
	defer wg.Done()

	r := newReadBuffer(c.opts.bufSize, c.opts.maxSize, c.pool, p)
	for {
		if ok := r.read(nc); !ok {
			r.finalize()
			break
		}
	}
}
