package client

import "time"

func (c *connection) terminator() {
	defer c.w.Done()

	ch := c.c.opts.termFlushCh
	if ch == nil {
		t := time.NewTicker(c.c.opts.termInt)
		defer t.Stop()

		ch = t.C
	}

	for {
		select {
		case <-c.t:
			c.p.flush()
			return

		case t := <-ch:
			if c.p.check(t) {
				c.closeNet()
			}
		}
	}
}
