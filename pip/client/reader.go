package client

func (c *connection) reader() {
	defer c.w.Done()

	r := newReadBuffer(c.c.opts.bufSize, c.c.opts.maxSize, c.c.pool, c.p)
	for {
		if err := r.read(c.n); err != nil {
			if !isConnClosed(err) {
				c.closeNet()
			}

			r.finalize()
			break
		}
	}
}
