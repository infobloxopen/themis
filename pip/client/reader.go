package client

func (c *connection) reader() {
	defer c.w.Done()

	r := newReadBuffer(c.c.opts.bufSize, c.c.opts.maxSize, c.c.pool, c.p)
	for {
		if ok := r.read(c.n); !ok {
			r.finalize()
			break
		}
	}
}
