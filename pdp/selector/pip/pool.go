package pip

import (
	"sync"
	"time"

	"github.com/infobloxopen/themis/pip/client"
)

type clientsPool struct {
	sync.RWMutex

	net string
	k8s bool
	m   map[string]timedClient
}

// NewTCPClientsPool creates PIP clients pool for "pip" selector schema.
func NewTCPClientsPool() *clientsPool {
	return &clientsPool{
		net: "tcp",
		m:   make(map[string]timedClient),
	}
}

// NewUnixClientsPool creates PIP clients pool for "pip+unix" selector schema.
func NewUnixClientsPool() *clientsPool {
	return &clientsPool{
		net: "unix",
		m:   make(map[string]timedClient),
	}
}

// NewK8sClientsPool creates PIP clients pool for "pip+k8s" selector schema.
func NewK8sClientsPool() *clientsPool {
	return &clientsPool{
		net: "tcp",
		k8s: true,
		m:   make(map[string]timedClient),
	}
}

func (p *clientsPool) cleaner(ch <-chan time.Time, done <-chan struct{}) {
	var t *time.Ticker
	if ch == nil {
		t = time.NewTicker(10 * time.Second)
		ch = t.C
	}

Loop:
	for {
		select {
		case _, ok := <-done:
			if !ok {
				break Loop
			}
		case now := <-ch:
			p.cleanupAll(now)
		}
	}

	if t != nil {
		t.Stop()
	}

	p.Lock()
	defer p.Unlock()
	for a, c := range p.m {
		delete(p.m, a)
		c.c.Close()
	}
}

func (p *clientsPool) getExpired(t int64) (string, timedClient) {
	p.RLock()
	defer p.RUnlock()

	for a, c := range p.m {
		if c.check(t) {
			return a, c
		}
	}

	return "", timedClient{}
}

func (p *clientsPool) cleanup(a string, c timedClient, t int64) client.Client {
	p.Lock()
	defer p.Unlock()

	if c.check(t) {
		delete(p.m, a)
		return c.c
	}

	return nil
}

func (p *clientsPool) cleanupAll(t time.Time) {
	tns := t.UnixNano()

	for {
		a, c := p.getExpired(tns)
		if len(a) <= 0 {
			break
		}

		if cc := p.cleanup(a, c, tns); cc != nil {
			cc.Close()
		}
	}
}

func (p *clientsPool) Get(addr string) (client.Client, error) {
	if c, ok := p.rawGet(addr); ok {
		return c, nil
	}

	return p.getOrNew(addr)
}

func (p *clientsPool) Free(addr string) {
	p.RLock()
	defer p.RUnlock()

	if c, ok := p.m[addr]; ok {
		c.free()
	}
}

func (p *clientsPool) rawGet(addr string) (client.Client, bool) {
	p.RLock()
	defer p.RUnlock()

	if c, ok := p.m[addr]; ok {
		return c.markAndGet(), true
	}

	return nil, false
}

func (p *clientsPool) getOrNew(addr string) (client.Client, error) {
	p.Lock()
	defer p.Unlock()

	if c, ok := p.m[addr]; ok {
		return c.markAndGet(), nil
	}

	c, err := makeTimedClient(p.net, addr, p.k8s)
	if err != nil {
		return nil, err
	}

	p.m[addr] = c
	return c.markAndGet(), nil
}
