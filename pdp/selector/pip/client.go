package pip

import (
	"sync/atomic"
	"time"

	"github.com/infobloxopen/themis/pip/client"
)

// SetClientTTL sets duration after which unused PIP client is closed.
func SetClientTTL(t time.Duration) {
	atomic.StoreInt64(clientTTL, t.Nanoseconds())
}

// SetHotSpotBalancer turns on hot-spot balancer for new PIP clients.
func SetHotSpotBalancer() {
	atomic.StoreInt64(isHotSpot, 1)
}

// SetRoundRobinBalancer turns on round-robin balancer for new PIP clients.
func SetRoundRobinBalancer() {
	atomic.StoreInt64(isHotSpot, 0)
}

// SetCacheWithTTL enables cache with given TTL for new PIP clients.
func SetCacheWithTTL(t time.Duration) {
	cacheOpts.Store(cacheOptions{
		cache: true,
		ttl:   t,
	})
}

// SetCacheWithTTLAndMaxSize enables cache with given TTL and size limit for
// new PIP clients.
func SetCacheWithTTLAndMaxSize(t time.Duration, s int) {
	cacheOpts.Store(cacheOptions{
		cache: true,
		ttl:   t,
		size:  s,
	})
}

// ClearCache turns off cache for new clients.
func ClearCache() {
	cacheOpts.Store(cacheOptions{
		cache: false,
	})
}

type timedClient struct {
	t *int64
	u *int64
	c client.Client
}

type cacheOptions struct {
	cache bool
	ttl   time.Duration
	size  int
}

var (
	clientTTL *int64
	isHotSpot *int64
	cacheOpts *atomic.Value
)

func init() {
	clientTTL = new(int64)
	SetClientTTL(time.Minute)

	isHotSpot = new(int64)
	SetHotSpotBalancer()

	cacheOpts = new(atomic.Value)
	ClearCache()
}

func makeClientOptions(net, addr string, k8s bool) []client.Option {
	return append(
		[]client.Option{
			client.WithNetwork(net),
			client.WithAddress(addr),
			makeBalancerOption(),
			makeRadarOption(k8s),
		},
		makeCacheOptions()...,
	)
}

func makeBalancerOption() client.Option {
	if atomic.LoadInt64(isHotSpot) != 0 {
		return client.WithHotSpotBalancer()
	}

	return client.WithRoundRobinBalancer()
}

func makeRadarOption(k8s bool) client.Option {
	if k8s {
		return client.WithK8sRadar()
	}

	return client.WithDNSRadar()
}

func makeCacheOptions() []client.Option {
	if v := cacheOpts.Load(); v != nil {
		if co := v.(cacheOptions); co.cache {
			if co.size > 0 {
				return []client.Option{client.WithCacheTTLAndMaxSize(co.ttl, co.size)}
			}

			return []client.Option{client.WithCacheTTL(co.ttl)}
		}
	}

	return nil
}

func makeTimedClient(net, addr string, k8s bool) (timedClient, error) {
	c := client.NewClient(makeClientOptions(net, addr, k8s)...)
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
	return t-ct > atomic.LoadInt64(clientTTL)
}
