package pip

import (
	"sync"

	"github.com/infobloxopen/themis/pip/client"
)

type clientsPool struct {
	sync.RWMutex

	net string
	k8s bool
	m   map[string]client.Client
}

var (
	pipClients = &clientsPool{
		net: "tcp",
	}

	pipUnixClients = &clientsPool{
		net: "unix",
	}

	pipK8sClients = &clientsPool{
		net: "tcp",
		k8s: true,
	}
)

func (p *clientsPool) Get(addr string) (client.Client, error) {
	if c, ok := p.rawGet(addr); ok {
		return c, nil
	}

	p.Lock()
	defer p.Unlock()

	opts := []client.Option{
		client.WithNetwork(p.net),
		client.WithAddress(addr),
		client.WithHotSpotBalancer(),
	}
	if p.k8s {
		opts = append(opts,
			client.WithK8sRadar(),
		)
	} else {
		opts = append(opts,
			client.WithDNSRadar(),
		)
	}

	c := client.NewClient(opts...)
	if err := c.Connect(); err != nil {
		return nil, err
	}

	p.m[addr] = c
	return c, nil
}

func (p *clientsPool) rawGet(addr string) (client.Client, bool) {
	p.RLock()
	defer p.RUnlock()

	c, ok := p.m[addr]
	return c, ok
}
