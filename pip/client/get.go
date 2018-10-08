package client

import "github.com/infobloxopen/themis/pdp"

func (c *client) Get(args ...pdp.AttributeAssignment) (pdp.AttributeValue, error) {
	conn := c.getConnection()
	if conn == nil {
		return pdp.UndefinedValue, ErrNotConnected
	}
	defer conn.g.Done()

	b := c.pool.Get()
	defer c.pool.Put(b)

	n, err := pdp.MarshalRequestAssignmentsToBuffer(b, args)
	if err != nil {
		return pdp.UndefinedValue, err
	}

	b, err = conn.get(b[:n])
	if b != nil {
		c.pool.Put(b[:cap(b)])
	}
	return pdp.UndefinedValue, err
}

func (c *client) getConnection() *connection {
	c.RLock()
	defer c.RUnlock()

	conn := c.c
	if conn != nil {
		conn.g.Add(1)
	}

	return conn
}

func (c *connection) get(b []byte) ([]byte, error) {
	i, p := c.p.alloc()
	defer c.p.free(i)

	c.r <- request{
		i: i,
		b: b,
	}

	return p.get()
}
