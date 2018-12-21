package client

import (
	"sync"
	"time"
)

type dNSRadar struct {
	sync.Mutex

	done chan struct{}
	t    *time.Ticker

	addr string
	d    time.Duration
}

func newDNSRadar(addr string, d time.Duration) *dNSRadar {
	return &dNSRadar{
		addr: addr,
		d:    d,
		done: make(chan struct{}),
	}
}

func (r *dNSRadar) start(addrs []string) <-chan addrUpdate {
	r.Lock()
	defer r.Unlock()

	if r.t != nil || r.done == nil {
		return nil
	}
	r.t = time.NewTicker(r.d)

	ch := make(chan addrUpdate, 1024)
	go runDNSRadar(r.done, ch, r.t.C, r.addr, addrs)

	return ch
}

func (r *dNSRadar) stop() {
	r.Lock()
	defer r.Unlock()

	if r.done == nil {
		return
	}

	if r.t != nil {
		r.t.Stop()
		r.t = nil
	}

	close(r.done)
	r.done = nil
}

func runDNSRadar(done <-chan struct{}, ch chan addrUpdate, t <-chan time.Time, addr string, addrs []string) {
	defer close(ch)

	idx := make(map[string]struct{})
	for _, a := range addrs {
		idx[a] = struct{}{}
	}

	for {
		select {
		case _, ok := <-done:
			if !ok {
				return
			}

		case <-t:
			idx = lookupDNSRadar(ch, idx, addr)
		}
	}
}

func lookupDNSRadar(ch chan addrUpdate, idx map[string]struct{}, addr string) map[string]struct{} {
	addrs, err := lookupHostPort(addr)
	if err != nil {
		ch <- addrUpdate{err: err}
		return idx
	}

	return dispatchDNSRadar(ch, idx, addrs)
}

func dispatchDNSRadar(ch chan addrUpdate, idx map[string]struct{}, addrs []string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, a := range addrs {
		out[a] = struct{}{}

		if _, ok := idx[a]; !ok {
			ch <- addrUpdate{
				op:   addrUpdateOpAdd,
				addr: a,
			}
		}
	}

	for a := range idx {
		if _, ok := out[a]; !ok {
			ch <- addrUpdate{
				op:   addrUpdateOpDel,
				addr: a,
			}
		}
	}

	return out
}
