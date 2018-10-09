package client

import "time"

func (c *connection) writer() {
	defer c.w.Done()

	w := newWriteBuffer(c.n, c.c.opts.bufSize, c.p)

	ch := c.c.opts.writeFlushCh
	if ch == nil {
		t := time.NewTicker(c.c.opts.writeInt)
		defer t.Stop()

		ch = t.C
	}

	for {
		select {
		case r, ok := <-c.r:
			if !ok {
				w.flush()
				return
			}

			w.put(r)
			c.c.pool.Put(r.b)

		case <-ch:
			w.flush()
		}
	}
}
