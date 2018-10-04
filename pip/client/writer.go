package client

import (
	"net"
	"sync"
	"time"
)

func (c *client) writer(wg *sync.WaitGroup, nc net.Conn, req chan request, p pipes, inc chan int) {
	defer wg.Done()

	w := newWriteBuffer(nc, c.opts.bufSize, p, inc)

	ch := c.opts.writeFlushCh
	if ch == nil {
		t := time.NewTicker(c.opts.writeInt)
		defer t.Stop()

		ch = t.C
	}

	for {
		select {
		case r, ok := <-req:
			if !ok {
				w.finalize()
				return
			}

			w.put(r)

		case <-ch:
			w.flush()
		}
	}
}
