package pip

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/infobloxopen/themis/pdp"
)

func TestNewTCPClientsPool(t *testing.T) {
	pc := NewTCPClientsPool()
	if pc.net != "tcp" {
		t.Errorf("expected %q network but got %q", "tcp", pc.net)
	}

	if pc.m == nil {
		t.Error("expected initialized map")
	}

	if pc.k8s {
		t.Error("expected no kubernetes")
	}
}

func TestNewUnixClientsPool(t *testing.T) {
	pc := NewUnixClientsPool()
	if pc.net != "unix" {
		t.Errorf("expected %q network but got %q", "unix", pc.net)
	}

	if pc.m == nil {
		t.Error("expected initialized map")
	}

	if pc.k8s {
		t.Error("expected no kubernetes")
	}
}

func TestNewK8sClientsPool(t *testing.T) {
	pc := NewK8sClientsPool()
	if pc.net != "tcp" {
		t.Errorf("expected %q network but got %q", "tcp", pc.net)
	}

	if pc.m == nil {
		t.Error("expected initialized map")
	}

	if !pc.k8s {
		t.Error("expected kubernetes")
	}
}

func TestClientsPoolCleaner(t *testing.T) {
	ccExp := new(testPipClient)
	tcExp := timedClient{
		t: new(int64),
		u: new(int64),
		c: ccExp,
	}
	tcExp.markAndGet()
	tcExp.free()

	cc := new(testPipClient)
	tc := timedClient{
		t: new(int64),
		u: new(int64),
		c: cc,
	}
	tc.markAndGet()
	tc.free()

	pc := NewTCPClientsPool()
	pc.Lock()
	pc.m["127.0.0.1:5600"] = tcExp
	pc.m["127.0.0.2:5600"] = tc
	pc.Unlock()

	ch := make(chan time.Time)
	done := make(chan struct{})

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()

		pc.cleaner(ch, done)
	}()

	tc.markAndGet()
	oldTime := atomic.LoadInt64(tc.t)
	ch <- time.Unix(0, oldTime+(2*time.Minute).Nanoseconds())

	time.Sleep(100 * time.Millisecond)

	pc.Lock()
	count := len(pc.m)
	pc.Unlock()
	if count != 1 {
		t.Errorf("expected %d in clients map but got %d", 1, count)
	}

	ccExp.RLock()
	closed := ccExp.closed
	ccExp.RUnlock()
	if !closed {
		t.Error("expected closed client")
	}

	cc.RLock()
	closed = cc.closed
	cc.RUnlock()
	if closed {
		t.Error("expected not closed client")
	}

	tc.free()

	close(done)
	wg.Wait()

	pc.Lock()
	count = len(pc.m)
	pc.Unlock()
	if count != 0 {
		t.Errorf("expected %d in clients map but got %d", 0, count)
	}

	cc.RLock()
	closed = cc.closed
	cc.RUnlock()
	if !closed {
		t.Error("expected closed client")
	}
}

func TestClientsPoolCleanerWithInternalTicker(t *testing.T) {
	cc := new(testPipClient)
	tc := timedClient{
		t: new(int64),
		u: new(int64),
		c: cc,
	}
	tc.markAndGet()
	tc.free()

	pc := NewTCPClientsPool()
	pc.Lock()
	pc.m["localhost:5600"] = tc
	pc.Unlock()

	done := make(chan struct{})

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()

		pc.cleaner(nil, done)
	}()

	close(done)
	wg.Wait()

	pc.Lock()
	count := len(pc.m)
	pc.Unlock()
	if count != 0 {
		t.Errorf("expected %d in clients map but got %d", 0, count)
	}

	cc.RLock()
	closed := cc.closed
	cc.RUnlock()
	if !closed {
		t.Error("expected closed client")
	}
}

func TestClientsPoolGetExpired(t *testing.T) {
	tc := timedClient{
		t: new(int64),
		u: new(int64),
		c: new(testPipClient),
	}
	tc.markAndGet()
	tc.free()

	pc := NewTCPClientsPool()
	pc.Lock()
	pc.m["localhost:5600"] = tc
	pc.Unlock()

	oldTime := atomic.LoadInt64(tc.t)

	a, c := pc.getExpired(oldTime)
	if a != "" {
		t.Errorf("expected no address but got %q", a)
	}
	if c.c != nil {
		t.Errorf("expected no client but got %#v", c.c)
	}

	a, c = pc.getExpired(oldTime + (2 * time.Minute).Nanoseconds())
	if a != "localhost:5600" {
		t.Errorf("expected %q address but got %q", "localhost:5600", a)
	}
	if c.c != tc.c {
		t.Errorf("expected %#v client but got %#v", tc.c, c.c)
	}
}

func TestClientsPoolCleanup(t *testing.T) {
	tc := timedClient{
		t: new(int64),
		u: new(int64),
		c: new(testPipClient),
	}
	tc.markAndGet()
	tc.free()

	pc := NewTCPClientsPool()
	pc.Lock()
	pc.m["localhost:5600"] = tc
	pc.Unlock()

	oldTime := atomic.LoadInt64(tc.t)

	pc.cleanup("localhost:5600", tc, oldTime)
	pc.Lock()
	count := len(pc.m)
	pc.Unlock()
	if count != 1 {
		t.Errorf("expected %d in clients map but got %d", 1, count)

	}

	pc.cleanup("localhost:5600", tc, oldTime+(2*time.Minute).Nanoseconds())
	pc.Lock()
	count = len(pc.m)
	pc.Unlock()
	if count != 0 {
		t.Errorf("expected %d in clients map but got %d", 0, count)

	}
}

func TestClientsPoolCleanupAll(t *testing.T) {
	cc := new(testPipClient)
	tc := timedClient{
		t: new(int64),
		u: new(int64),
		c: cc,
	}
	tc.markAndGet()
	tc.free()

	pc := NewTCPClientsPool()
	pc.Lock()
	pc.m["localhost:5600"] = tc
	pc.Unlock()

	oldTime := atomic.LoadInt64(tc.t)

	pc.cleanupAll(time.Unix(0, oldTime))
	pc.Lock()
	count := len(pc.m)
	pc.Unlock()
	if count != 1 {
		t.Errorf("expected %d in clients map but got %d", 1, count)

	}

	cc.RLock()
	closed := cc.closed
	cc.RUnlock()
	if closed {
		t.Error("expected not closed client")
	}

	pc.cleanupAll(time.Unix(0, oldTime+(2*time.Minute).Nanoseconds()))
	pc.Lock()
	count = len(pc.m)
	pc.Unlock()
	if count != 0 {
		t.Errorf("expected %d in clients map but got %d", 0, count)

	}

	cc.RLock()
	closed = cc.closed
	cc.RUnlock()
	if !closed {
		t.Error("expected closed client")
	}
}

func TestClientsPoolGet(t *testing.T) {
	pc := NewTCPClientsPool()
	c0, err := pc.Get("localhost:5600")
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	}
	if c0 == nil {
		t.Error("expected instance of client.Client but got nothing")
	} else {
		defer c0.Close()

		c1, err := pc.Get("localhost:5600")
		if err != nil {
			t.Errorf("expected no error but got %#v", err)
		}
		if c0 != c1 {
			t.Errorf("expected the same client %#v but got %#v", c0, c1)
		}
	}
}

func TestClientsPoolFree(t *testing.T) {
	tc := timedClient{
		t: new(int64),
		u: new(int64),
		c: new(testPipClient),
	}
	tc.markAndGet()
	tc.free()

	pc := NewTCPClientsPool()
	pc.Lock()
	pc.m["localhost:5600"] = tc
	pc.Unlock()

	_, err := pc.Get("localhost:5600")
	if err != nil {
		t.Fatal(err)
	}

	if count := atomic.LoadInt64(tc.u); count != 1 {
		t.Errorf("expected %d mark but got %d", 1, count)
	}

	pc.Free("localhost:5600")
	if count := atomic.LoadInt64(tc.u); count != 0 {
		t.Errorf("expected %d mark but got %d", 0, count)
	}
}

func TestClientsPoolRawGet(t *testing.T) {
	tc := timedClient{
		t: new(int64),
		u: new(int64),
		c: new(testPipClient),
	}
	tc.markAndGet()
	tc.free()

	pc := NewTCPClientsPool()

	pc.Lock()
	pc.m["localhost:5600"] = tc
	pc.Unlock()

	c, ok := pc.rawGet("localhost:5600")
	if !ok {
		t.Errorf("expected true but got %#v", ok)
	}
	if c != tc.c {
		t.Errorf("expected instance (%#v) of testPipClient but got %#v", tc.c, c)
	}

	c, ok = pc.rawGet("127.0.0.1:5600")
	if ok {
		t.Errorf("expected false but got %#v", ok)
	}
	if c != nil {
		t.Errorf("expected no testPipClient but got %#v", c)
	}
}

func TestClientsPoolGetOrNew(t *testing.T) {
	tc := timedClient{
		t: new(int64),
		u: new(int64),
		c: new(testPipClient),
	}
	tc.markAndGet()
	tc.free()

	pc := NewTCPClientsPool()

	pc.Lock()
	pc.m["localhost:5600"] = tc
	pc.Unlock()

	c, err := pc.getOrNew("localhost:5600")
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	}
	if c != tc.c {
		t.Errorf("expected instance (%#v) of testPipClient but got %#v", tc.c, c)
	}

	c, err = pc.getOrNew("127.0.0.1:5600")
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	}
	if c == nil {
		t.Error("expected instance of client.Client but got nothing")
	} else {
		c.Close()
	}

	kc := NewK8sClientsPool()
	c, err = kc.getOrNew("value.key.namespace:5600")
	if err == nil {
		t.Error("expected error")
	}
	if c != nil {
		t.Errorf("expected no client but got %#v", c)
		c.Close()
	}
}

type testPipClient struct {
	sync.RWMutex

	closed bool
}

func (c *testPipClient) Connect() error { panic("not implemented") }
func (c *testPipClient) Close() {
	c.Lock()
	defer c.Unlock()

	c.closed = true
}

func (c *testPipClient) Get(string, []pdp.AttributeValue) (pdp.AttributeValue, error) {
	panic("not implemented")
}
