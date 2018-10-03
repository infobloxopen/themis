package client

import "github.com/infobloxopen/themis/pdp"

func (c *client) Get(args ...pdp.AttributeAssignment) (pdp.AttributeValue, error) {
	c.lock.RLock()
	if c.gwg == nil {
		c.lock.RUnlock()
		return pdp.UndefinedValue, ErrNotConnected
	}

	c.gwg.Add(1)
	defer c.gwg.Done()

	req := c.req
	ps := c.pipes
	c.lock.RUnlock()

	b := c.pool.Get()
	defer c.pool.Put(b)

	n, err := pdp.MarshalRequestAssignmentsToBuffer(b, args)
	if err != nil {
		return pdp.UndefinedValue, err
	}

	i, p := ps.alloc()
	defer ps.free(i)

	req <- request{
		i: i,
		b: b[:n],
	}

	b, err = p.get()
	if b != nil {
		c.pool.Put(b[:cap(b)])
	}
	return pdp.UndefinedValue, err
}
