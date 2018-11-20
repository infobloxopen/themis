package pip

import (
	"sync/atomic"
	"testing"
	"time"
)

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
