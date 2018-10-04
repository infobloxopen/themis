package client

import (
	"net"
	"sync"
)

func (c *client) reader(wg *sync.WaitGroup, nc net.Conn, p pipes, dec chan int) {
	defer wg.Done()

	r := newReadBuffer(c.opts.bufSize, c.opts.maxSize, c.pool, p, dec)
	for {
		if ok := r.read(nc); !ok {
			r.finalize()
			break
		}
	}
}
