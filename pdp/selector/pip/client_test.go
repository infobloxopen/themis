package pip

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSetClientTTL(t *testing.T) {
	defer SetClientTTL(time.Minute)

	SetClientTTL(time.Hour)
	if d := atomic.LoadInt64(clientTTL); d != time.Hour.Nanoseconds() {
		t.Errorf("expected %s TTL but got %s", time.Hour, time.Duration(d))
	}
}

func TestSetBalancer(t *testing.T) {
	defer SetHotSpotBalancer()

	SetRoundRobinBalancer()
	if b := atomic.LoadInt64(isHotSpot); b != 0 {
		t.Errorf("expected %d but got %d", 0, b)
	}

	SetHotSpotBalancer()
	if b := atomic.LoadInt64(isHotSpot); b != 1 {
		t.Errorf("expected %d but got %d", 1, b)
	}
}

func TestCacheOptions(t *testing.T) {
	defer ClearCache()

	SetCacheWithTTL(time.Hour)
	v := cacheOpts.Load()
	if v == nil {
		t.Error("expected cache options but got nothing")
	} else if co, ok := v.(cacheOptions); ok {
		if !co.cache {
			t.Errorf("expected cache enabled but got %#v", co.cache)
		}

		if co.ttl != time.Hour {
			t.Errorf("expected %s TTL but got %s", time.Hour, co.ttl)
		}

		if co.size != 0 {
			t.Errorf("expected no size limit but got %d", co.size)
		}
	} else {
		t.Errorf("expected cacheOptions but got %T (%#v)", v, v)
	}

	SetCacheWithTTLAndMaxSize(time.Minute, 1024*1024)
	v = cacheOpts.Load()
	if v == nil {
		t.Error("expected cache options but got nothing")
	} else if co, ok := v.(cacheOptions); ok {
		if !co.cache {
			t.Errorf("expected cache enabled but got %#v", co.cache)
		}

		if co.ttl != time.Minute {
			t.Errorf("expected %s TTL but got %s", time.Minute, co.ttl)
		}

		if co.size != 1024*1024 {
			t.Errorf("expected %d size limit but got %d", 1024*1024, co.size)
		}
	} else {
		t.Errorf("expected cacheOptions but got %T (%#v)", v, v)
	}

	ClearCache()
	v = cacheOpts.Load()
	if v == nil {
		t.Error("expected cache options but got nothing")
	} else if co, ok := v.(cacheOptions); ok {
		if co.cache {
			t.Errorf("expected cache disabled but got %#v", co.cache)
		}

		if co.ttl != 0 {
			t.Errorf("expected %s TTL but got %s", time.Duration(0), co.ttl)
		}

		if co.size != 0 {
			t.Errorf("expected no size limit but got %d", co.size)
		}
	} else {
		t.Errorf("expected cacheOptions but got %T (%#v)", v, v)
	}
}

func TestMakeClientOptions(t *testing.T) {
	opts := makeClientOptions("tcp", "localhost:5600", false)
	if len(opts) <= 0 {
		t.Errorf("expected some options but got %#v", opts)
	}
}

func TestMakeBalancerOption(t *testing.T) {
	defer SetHotSpotBalancer()

	SetRoundRobinBalancer()
	if opt := makeBalancerOption(); opt == nil {
		t.Error("expected some option")
	}

	SetHotSpotBalancer()
	if opt := makeBalancerOption(); opt == nil {
		t.Error("expected some option")
	}
}

func TestMakeRadarOption(t *testing.T) {
	if opt := makeRadarOption(true); opt == nil {
		t.Error("expected some option")
	}

	if opt := makeRadarOption(false); opt == nil {
		t.Error("expected some option")
	}
}

func TestMakeCacheOptions(t *testing.T) {
	defer ClearCache()

	SetCacheWithTTL(time.Minute)
	if opts := makeCacheOptions(); len(opts) != 1 {
		t.Errorf("expected an option but got %#v", opts)
	}

	SetCacheWithTTLAndMaxSize(time.Minute, 1024*1024)
	if opts := makeCacheOptions(); len(opts) != 1 {
		t.Errorf("expected an option but got %#v", opts)
	}

	ClearCache()
	if opts := makeCacheOptions(); len(opts) != 0 {
		t.Errorf("expected no options but got %#v", opts)
	}
}

func TestMakeTimedClient(t *testing.T) {
	c, err := makeTimedClient("tcp", "localhost:5600", false)
	if err != nil {
		t.Error(err)
	} else {
		if c.c == nil {
			t.Error("expected some client but got nothing")
		} else {
			defer c.c.Close()
		}

		if c.t == nil {
			t.Error("expected pointer to timestamp but got nothing")
		}
	}

	c, err = makeTimedClient("tcp", "value.key.namespace:5600", true)
	if err == nil {
		t.Errorf("expected error but got client %#v", c)
	}
}

func TestTimedClientMarkAndGet(t *testing.T) {
	c, err := makeTimedClient("tcp", "localhost:5600", false)
	if err != nil {
		t.Fatal(err)
	}
	defer c.c.Close()

	oldTime := time.Now().UnixNano() - 100
	atomic.StoreInt64(c.t, oldTime)

	cc := c.markAndGet()

	newTime := atomic.LoadInt64(c.t)
	if newTime <= oldTime {
		t.Errorf("expected new timestamp %d greater than old one %d", newTime, oldTime)
	}

	if cc != c.c {
		t.Errorf("expected %#v client but got %#v", c.c, cc)
	}

	count := atomic.LoadInt64(c.u)
	if count != 1 {
		t.Errorf("expected %d mark but got %d", 1, count)
	}
}

func TestTimedClientFree(t *testing.T) {
	c, err := makeTimedClient("tcp", "localhost:5600", false)
	if err != nil {
		t.Fatal(err)
	}
	defer c.c.Close()

	c.markAndGet()

	count := atomic.LoadInt64(c.u)
	if count != 1 {
		t.Errorf("expected %d mark but got %d", 1, count)
	}

	c.free()

	count = atomic.LoadInt64(c.u)
	if count != 0 {
		t.Errorf("expected %d mark but got %d", 0, count)
	}
}

func TestTimedClientCheck(t *testing.T) {
	c, err := makeTimedClient("tcp", "localhost:5600", false)
	if err != nil {
		t.Fatal(err)
	}
	defer c.c.Close()

	c.markAndGet()

	oldTime := atomic.LoadInt64(c.t)
	if exp := c.check(oldTime); exp {
		t.Error("expected not expired")
	}
	if exp := c.check(oldTime + (2 * time.Minute).Nanoseconds()); exp {
		t.Error("expected not expired")
	}

	c.free()

	oldTime = atomic.LoadInt64(c.t)
	if exp := c.check(oldTime); exp {
		t.Error("expected not expired")
	}
	if exp := c.check(oldTime + (2 * time.Minute).Nanoseconds()); !exp {
		t.Error("expected expired")
	}
}
