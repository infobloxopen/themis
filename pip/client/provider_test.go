package client

import (
	"io"
	"net"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProviderStartStop(t *testing.T) {
	c := NewClient().(*client)
	c.d = testProviderDialer{}
	p := new(provider)

	p.start(c, []string{"127.0.0.1:5600"})
	p.RLock()
	if assert.True(t, p.started) {
		wg := p.wg
		assert.NotZero(t, wg)

		rCnd := p.rCnd
		assert.NotZero(t, rCnd)

		wCnd := p.wCnd
		assert.NotZero(t, wCnd)

		assert.Equal(t, c, p.c)

		idx := p.idx
		assert.NotZero(t, idx)

		queue := p.queue
		assert.NotZero(t, queue)

		healthy := p.healthy
		assert.NotZero(t, healthy)

		broken := p.broken
		assert.NotZero(t, broken)

		getter := p.getter
		assert.NotZero(t, getter)
		p.RUnlock()

		c1 := NewClient().(*client)
		p.start(c1, []string{"127.0.0.2:5600"})

		time.Sleep(time.Millisecond)

		p.RLock()
		assert.Equal(t, wg, p.wg)
		assert.Equal(t, rCnd, p.rCnd)
		assert.Equal(t, wCnd, p.wCnd)
		assert.Equal(t, c, p.c)
		assert.Equal(t, idx, p.idx)
		assert.Equal(t, queue, p.queue)
		assert.Equal(t, healthy, p.healthy)
		assert.Equal(t, broken, p.broken)
		assert.Equal(t, reflect.ValueOf(getter).Pointer(), reflect.ValueOf(p.getter).Pointer())
		p.RUnlock()

		p.stop()

		p.RLock()
		if assert.False(t, p.started) {

			assert.Zero(t, p.wg)
			assert.Zero(t, p.rCnd)
			assert.Zero(t, p.wCnd)
			assert.Zero(t, p.c)
			assert.Zero(t, p.idx)
			assert.Zero(t, p.queue)
			assert.Zero(t, p.healthy)
			assert.Zero(t, p.broken)
			assert.Zero(t, p.getter)
			p.RUnlock()

			p.stop()

			p.RLock()
			assert.False(t, p.started)
			p.RUnlock()
		} else {
			p.RUnlock()
		}
	} else {
		p.RUnlock()
	}
}

func TestProviderStopAndDisconnectAll(t *testing.T) {
	c := NewClient().(*client)
	d := newTestProviderSyncDialer()
	c.d = d

	c.Connect()
	defer c.Close()

	conn := <-d.ch

	bConn := newTestProviderConn("127.0.0.2:5600")
	cc := c.newConnection(bConn)
	cc.start()

	c.p.broken[cc.i] = destConn{c: cc}

	c.p.stop()

	assert.True(t, conn.isClosed())
	assert.True(t, bConn.isClosed())
}

func TestProviderGet(t *testing.T) {
	c := NewClient().(*client)
	d := newTestProviderSyncDialer()
	c.d = d

	c.Connect()
	defer c.Close()

	var pConn *connection
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()

		pConn = c.p.get()
	}()

	for c.p.rCnd == nil {
		time.Sleep(time.Microsecond)
	}
	c.p.rCnd.Broadcast()

	time.Sleep(time.Millisecond)

	conn := <-d.ch
	wg.Wait()

	if assert.NotZero(t, pConn) {
		assert.Equal(t, conn, pConn.n)
		pConn.g.Done()
	}

	c.p.stop()

	assert.True(t, conn.isClosed())

	assert.Zero(t, c.p.get())
}

func TestProviderConnector(t *testing.T) {
	c := NewClient().(*client)
	d := newTestProviderSyncDialer()
	c.d = d

	c.Connect()
	defer c.Close()

	conn := <-d.ch
	assert.Equal(t, "localhost:5600", conn.a)

	bConn := newTestProviderConn("127.0.0.2:5600")
	cc := c.newConnection(bConn)
	cc.start()

	c.p.Lock()
	c.p.broken[cc.i] = destConn{
		d: "127.0.0.3:5600",
		c: cc,
	}
	c.p.Unlock()
	c.p.wCnd.Signal()

	<-bConn.ch
	assert.True(t, bConn.isClosed())
	assert.Empty(t, c.p.broken)

	conn = <-d.ch
	assert.Equal(t, "127.0.0.3:5600", conn.a)
}

func TestProviderConnect(t *testing.T) {
	c := NewClient().(*client)
	d := newTestProviderSyncDialer()
	c.d = d
	atomic.StoreUint32(c.state, pipClientConnected)

	c.p.Lock()
	c.p.c = c
	c.p.started = true
	c.p.healthy = make(map[string]*connection)
	c.p.queue = []*connection{}
	c.p.rCnd = sync.NewCond(c.p.RLocker())
	c.p.Unlock()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go c.p.connect(wg, c, "127.0.0.1:5600")

	conn := <-d.ch
	wg.Wait()

	c.p.Lock()
	if assert.NotEmpty(t, c.p.healthy) {
		if hConn := c.p.healthy["127.0.0.1:5600"]; assert.NotZero(t, hConn) {
			assert.Equal(t, conn, hConn.n)
		}

		c.p.healthy = make(map[string]*connection)
	}

	if assert.NotEmpty(t, c.p.queue) {
		if qConn := c.p.queue[0]; assert.NotZero(t, qConn) {
			assert.Equal(t, conn, qConn.n)
		}

		c.p.queue = []*connection{}
	}

	assert.False(t, conn.isClosed())
	c.p.Unlock()

	c.p.started = false
	wg.Add(1)
	go c.p.connect(wg, c, "127.0.0.1:5600")

	conn = <-d.ch
	wg.Wait()

	c.p.RLock()
	assert.Empty(t, c.p.healthy)
	assert.Empty(t, c.p.queue)
	assert.True(t, conn.isClosed())
	c.p.RUnlock()
}

func TestProviderGetBroken(t *testing.T) {
	c := NewClient().(*client)

	c.p.Lock()
	c.p.c = c
	c.p.started = true
	c.p.wCnd = sync.NewCond(c.p)
	c.p.broken = make(map[uint64]destConn)
	c.p.Unlock()

	var d string
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()

		d, _ = c.p.getBroken()
	}()

	c.p.wCnd.Signal()
	time.Sleep(time.Millisecond)

	c.p.Lock()
	c.p.broken[c.nextID()] = destConn{d: "test"}
	c.p.wCnd.Signal()
	c.p.Unlock()

	wg.Wait()

	c.p.Lock()
	assert.Equal(t, "test", d)
	assert.Empty(t, c.p.broken)
	c.p.started = false
	c.p.Unlock()

	d, conn := c.p.getBroken()
	assert.Zero(t, d)
	assert.Zero(t, conn)
}

func TestProviderSelectGetter(t *testing.T) {
	g := selectGetter(balancerTypeSimple)
	assert.Equal(t, reflect.ValueOf(simpleGetter).Pointer(), reflect.ValueOf(g).Pointer())

	g = selectGetter(balancerTypeRoundRobin)
	assert.Equal(t, reflect.ValueOf(roundRobinGetter).Pointer(), reflect.ValueOf(g).Pointer())

	g = selectGetter(balancerTypeHotSpot)
	assert.Equal(t, reflect.ValueOf(hotSpotGetter).Pointer(), reflect.ValueOf(g).Pointer())

}

func TestProviderSimpleGetter(t *testing.T) {
	c := NewClient().(*client)
	n := newTestProviderConn("test")
	cc := c.newConnection(n)

	i := new(uint64)
	q := []*connection{cc}

	gc := simpleGetter(i, q)
	assert.Zero(t, *i)
	assert.Equal(t, cc, gc)

	gc = simpleGetter(i, q)
	assert.Zero(t, *i)
	assert.Equal(t, cc, gc)
}

func TestProviderRoundRobinGetter(t *testing.T) {
	c := NewClient().(*client)

	n1 := newTestProviderConn("test 1")
	cc1 := c.newConnection(n1)

	n2 := newTestProviderConn("test 2")
	cc2 := c.newConnection(n2)

	i := new(uint64)
	q := []*connection{cc1, cc2}

	gc := roundRobinGetter(i, q)
	assert.Equal(t, uint64(1), *i)
	assert.Equal(t, cc1, gc)

	gc = roundRobinGetter(i, q)
	assert.Equal(t, uint64(2), *i)
	assert.Equal(t, cc2, gc)

	gc = roundRobinGetter(i, q)
	assert.Equal(t, uint64(3), *i)
	assert.Equal(t, cc1, gc)
}

func TestProviderHotSpotGetter(t *testing.T) {
	c := NewClient(
		WithMaxQueue(1),
	).(*client)

	n1 := newTestProviderConn("test 1")
	cc1 := c.newConnection(n1)

	n2 := newTestProviderConn("test 2")
	cc2 := c.newConnection(n2)

	i := new(uint64)
	q := []*connection{cc1, cc2}

	gc := hotSpotGetter(i, q)
	assert.Equal(t, uint64(0), *i)
	assert.Equal(t, cc1, gc)

	gc = hotSpotGetter(i, q)
	assert.Equal(t, uint64(0), *i)
	assert.Equal(t, cc1, gc)

	cc1.r <- request{}

	gc = hotSpotGetter(i, q)
	assert.Equal(t, uint64(1), *i)
	assert.Equal(t, cc2, gc)

	cc2.r <- request{}

	gc = hotSpotGetter(i, q)
	assert.Equal(t, uint64(3), *i)
	assert.Equal(t, cc2, gc)
}

type testProviderDialer struct{}

func (d testProviderDialer) dial(a string) (net.Conn, error) {
	return newTestProviderConn(a), nil
}

type testProviderSyncDialer struct {
	ch chan *testProviderConn
}

func newTestProviderSyncDialer() testProviderSyncDialer {
	return testProviderSyncDialer{
		ch: make(chan *testProviderConn),
	}
}

func (d testProviderSyncDialer) dial(a string) (net.Conn, error) {
	c := newTestProviderConn(a)
	d.ch <- c
	return c, nil
}

type testProviderConn struct {
	a      string
	closed *uint32
	ch     chan struct{}
}

func newTestProviderConn(a string) *testProviderConn {
	return &testProviderConn{
		a:      a,
		closed: new(uint32),
		ch:     make(chan struct{}),
	}
}

func (c *testProviderConn) isClosed() bool {
	return atomic.LoadUint32(c.closed) != 0
}

func (c *testProviderConn) Close() error {
	if atomic.CompareAndSwapUint32(c.closed, 0, 1) {
		close(c.ch)
	}

	return nil
}

func (c *testProviderConn) Write(b []byte) (int, error) {
	return len(b), nil
}

func (c *testProviderConn) Read(b []byte) (int, error) {
	<-c.ch
	return 0, io.EOF
}

func (c *testProviderConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c *testProviderConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c *testProviderConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c *testProviderConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c *testProviderConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
