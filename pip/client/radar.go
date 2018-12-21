package client

import "strings"

type radar interface {
	start(addrs []string) <-chan addrUpdate
	stop()
}

const (
	addrUpdateOpAdd = iota
	addrUpdateOpDel
)

type addrUpdate struct {
	op   int
	addr string
	err  error
}

func (c *client) newAddressesAndRadar() ([]string, radar, error) {
	if strings.ToLower(c.opts.net) == unixNet || c.opts.balancer == balancerTypeSimple {
		return []string{c.opts.addr}, nil, nil
	}

	if r, err := c.newRadar(); r != nil || err != nil {
		return c.opts.addrs, r, err
	}

	if c.opts.addrs == nil {
		addrs, err := lookupHostPort(c.opts.addr)
		return addrs, nil, err
	}

	return c.opts.addrs, nil, nil
}

func (c *client) newRadar() (radar, error) {
	switch c.opts.radar {
	case radarDNS:
		return newDNSRadar(c.opts.addr, c.opts.radarInt), nil

	case radarK8s:
		kc, err := c.opts.k8sClientMaker()
		if err != nil {
			return nil, err
		}

		r, err := newK8sRadar(c.opts.addr, kc, c.opts.radarInt)
		if err != nil {
			return nil, err
		}

		return r, nil
	}

	return nil, nil
}
