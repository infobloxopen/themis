package client

import "github.com/infobloxopen/themis/pdp"

func (c *client) Get(args ...pdp.AttributeAssignment) (pdp.AttributeValue, error) {
	c.lock.RLock()
	if c.rwg == nil {
		c.lock.RUnlock()
		return pdp.UndefinedValue, ErrNotConnected
	}

	c.rwg.Add(1)
	req := c.req
	ps := c.pipes
	c.lock.RUnlock()
	defer c.rwg.Done()

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

	_, err = p.get()
	return pdp.UndefinedValue, err
}
