package client

import (
	"sync"
	"time"
)

func (c *client) writer(wg *sync.WaitGroup, req chan request, p pipes) {
	defer wg.Done()

	w := newWriteBuffer(c.opts.bufSize, p)

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
				if !w.isEmpty() {
					w.flush()
				}

				return
			}

			w.put(r)

		case <-ch:
			if !w.isEmpty() {
				w.flush()
			}
		}
	}
}
