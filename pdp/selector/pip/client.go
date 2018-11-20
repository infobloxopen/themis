package pip

import (
	"sync/atomic"
	"time"

	"github.com/infobloxopen/themis/pip/client"
)

type timedClient struct {
	t *int64
	u *int64
	c client.Client
}

var clientTTL = time.Minute.Nanoseconds()

func makeTimedClient(net, addr string, k8s bool) (timedClient, error) {
	opts := []client.Option{
		client.WithNetwork(net),
		client.WithAddress(addr),
		client.WithHotSpotBalancer(),
	}
	if k8s {
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
		return timedClient{}, err
	}

	return timedClient{
		t: new(int64),
		u: new(int64),
		c: c,
	}, nil
}

func (c timedClient) markAndGet() client.Client {
	atomic.StoreInt64(c.t, time.Now().UnixNano())
	atomic.AddInt64(c.u, 1)
	return c.c
}

func (c timedClient) free() {
	atomic.StoreInt64(c.t, time.Now().UnixNano())
	atomic.AddInt64(c.u, -1)
}

func (c timedClient) check(t int64) bool {
	cu := atomic.LoadInt64(c.u)
	if cu > 0 {
		return false
	}

	ct := atomic.LoadInt64(c.t)
	return t-ct > clientTTL
}
