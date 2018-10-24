package client

import (
	"errors"
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
	r := new(testProviderRadar)
	c.r = r
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

		retry := p.retry
		assert.NotZero(t, retry)

		getter := p.getter
		assert.NotZero(t, getter)

		ch := r.ch
		assert.NotZero(t, ch)
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
		assert.Equal(t, retry, p.retry)
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

			assert.Zero(t, r.ch)
			select {
			default:
				assert.Fail(t, "radar channel hasn't been closed")
			case u, ok := <-ch:
				assert.False(t, ok, "update %#v", u)
			}

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

	ch := make(chan struct{})

	c.p.Lock()
	c.p.broken[cc.i] = destConn{c: cc}
	c.p.retry["127.0.0.2:5601"] = ch
	c.p.Unlock()

	c.p.stop()

	assert.True(t, conn.isClosed())
	assert.True(t, bConn.isClosed())
	_, ok := <-ch
	assert.False(t, ok)
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
	c.p.retry = make(map[string]chan struct{})
	c.p.queue = []*connection{}
	c.p.rCnd = sync.NewCond(c.p.RLocker())

	done := make(chan struct{})
	c.p.retry["127.0.0.1:5600"] = done
	c.p.Unlock()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go c.p.connect(wg, c, "127.0.0.1:5600", done)

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

	assert.Empty(t, c.p.retry)
	assert.False(t, conn.isClosed())

	done = make(chan struct{})
	c.p.started = false
	c.p.Unlock()

	wg.Add(1)
	go c.p.connect(wg, c, "127.0.0.1:5600", done)

	conn = <-d.ch
	wg.Wait()

	c.p.RLock()
	assert.Empty(t, c.p.healthy)
	assert.Empty(t, c.p.queue)
	assert.True(t, conn.isClosed())
	c.p.RUnlock()

	var (
		cAddr net.Addr
		cErr  error
	)
	bd := newTestProviderSyncDialerBrokenConn(errors.New("test"))
	c.d = bd
	c.opts.onErr = func(a net.Addr, err error) {
		cAddr = a
		cErr = err
	}

	wg.Add(1)
	go c.p.connect(wg, c, "127.0.0.1:5600", make(chan struct{}))

	<-bd.ch
	wg.Wait()

	assert.Equal(t, "127.0.0.1:5600", cAddr.String())
	assert.Equal(t, bd.err, cErr)
}

func TestProviderGetBroken(t *testing.T) {
	c := NewClient().(*client)

	c.p.Lock()
	c.p.c = c
	c.p.started = true
	c.p.wCnd = sync.NewCond(c.p)
	c.p.broken = make(map[uint64]destConn)
	c.p.retry = make(map[string]chan struct{})
	c.p.Unlock()

	var (
		d  string
		bc *connection
		ch <-chan struct{}
	)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()

		d, bc, ch = c.p.getBroken()
	}()

	c.p.wCnd.Signal()
	time.Sleep(time.Millisecond)

	c.p.Lock()
	c.p.broken[c.nextID()] = destConn{d: "test"}
	c.p.wCnd.Signal()
	c.p.Unlock()

	wg.Wait()

	c.p.RLock()
	assert.Equal(t, "test", d)
	assert.Zero(t, bc)
	assert.NotZero(t, ch)
	assert.Empty(t, c.p.broken)
	c.p.RUnlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		d, bc, ch = c.p.getBroken()
	}()

	c.p.wCnd.Signal()
	time.Sleep(time.Millisecond)

	conn := c.newConnection(newTestProviderConn("127.0.0.1:5600"))
	c.p.Lock()
	c.p.broken[conn.i] = destConn{c: conn}
	c.p.wCnd.Signal()
	c.p.Unlock()

	wg.Wait()

	c.p.RLock()
	assert.Zero(t, d)
	assert.Equal(t, conn, bc)
	assert.Zero(t, ch)
	assert.Empty(t, c.p.broken)
	c.p.RUnlock()

	c.p.Lock()
	c.p.started = false
	c.p.Unlock()

	d, bc, ch = c.p.getBroken()
	assert.Zero(t, d)
	assert.Zero(t, bc)
	assert.Zero(t, ch)
}

func TestProviderUnqueue(t *testing.T) {
	c := NewClient().(*client)

	c1 := c.newConnection(newTestProviderConn("127.0.0.1:5601"))
	c2 := c.newConnection(newTestProviderConn("127.0.0.1:5602"))
	c3 := c.newConnection(newTestProviderConn("127.0.0.1:5603"))

	c.p.queue = []*connection{c1, c2, c3}
	c.p.unqueue(c2)
	assert.Equal(t, []*connection{c1, c3}, c.p.queue)

	c.p.queue = []*connection{c1, c2, c3}
	c.p.unqueue(c1)
	assert.Equal(t, []*connection{c2, c3}, c.p.queue)

	c.p.queue = []*connection{c1, c2, c3}
	c.p.unqueue(c3)
	assert.Equal(t, []*connection{c1, c2}, c.p.queue)
}

func TestProviderUnhealthy(t *testing.T) {
	c := NewClient().(*client)

	c1 := c.newConnection(newTestProviderConn("127.0.0.1:5601"))
	c2 := c.newConnection(newTestProviderConn("127.0.0.1:5602"))
	c3 := c.newConnection(newTestProviderConn("127.0.0.1:5603"))

	c.p.healthy = map[string]*connection{
		"127.0.0.1:5601": c1,
		"127.0.0.1:5602": c2,
		"127.0.0.1:5603": c3,
	}

	d := c.p.unhealthy(c2)
	assert.Equal(t, "127.0.0.1:5602", d)
	assert.Equal(t,
		map[string]*connection{
			"127.0.0.1:5601": c1,
			"127.0.0.1:5603": c3,
		},
		c.p.healthy,
	)

	d = c.p.unhealthy(c2)
	assert.Zero(t, d)
}

func TestProviderReport(t *testing.T) {
	c := NewClient().(*client)

	c1 := c.newConnection(newTestProviderConn("127.0.0.1:5601"))
	c2 := c.newConnection(newTestProviderConn("127.0.0.1:5602"))
	c3 := c.newConnection(newTestProviderConn("127.0.0.1:5603"))

	c.p.wCnd = sync.NewCond(c.p)
	c.p.broken = make(map[uint64]destConn)
	c.p.queue = []*connection{c1, c2, c3}
	c.p.healthy = map[string]*connection{
		"127.0.0.1:5601": c1,
		"127.0.0.1:5602": c2,
		"127.0.0.1:5603": c3,
	}

	c.p.report(c2)
	assert.Empty(t, c.p.broken)
	assert.Equal(t, []*connection{c1, c2, c3}, c.p.queue)
	assert.Equal(t,
		map[string]*connection{
			"127.0.0.1:5601": c1,
			"127.0.0.1:5602": c2,
			"127.0.0.1:5603": c3,
		},
		c.p.healthy,
	)

	c.p.started = true
	c.p.report(c2)
	assert.Equal(t,
		map[uint64]destConn{
			2: {
				d: "127.0.0.1:5602",
				c: c2,
			},
		},
		c.p.broken,
	)
	assert.Equal(t, []*connection{c1, c3}, c.p.queue)
	assert.Equal(t,
		map[string]*connection{
			"127.0.0.1:5601": c1,
			"127.0.0.1:5603": c3,
		},
		c.p.healthy,
	)

	c.p.report(c2)
	assert.Equal(t,
		map[uint64]destConn{
			2: {
				d: "127.0.0.1:5602",
				c: c2,
			},
		},
		c.p.broken,
	)
	assert.Equal(t, []*connection{c1, c3}, c.p.queue)
	assert.Equal(t,
		map[string]*connection{
			"127.0.0.1:5601": c1,
			"127.0.0.1:5603": c3,
		},
		c.p.healthy,
	)
}

func TestProviderChanger(t *testing.T) {
	tErr := errors.New("test")
	errs := []error{}

	c := NewClient().(*client)
	i1 := c.nextID()
	conn := &connection{
		i: i1,
	}

	wg := new(sync.WaitGroup)
	ch := make(chan addrUpdate)

	p := new(provider)
	p.Lock()
	p.c = c
	p.started = true
	p.wCnd = sync.NewCond(p)
	p.queue = []*connection{conn}
	p.healthy = map[string]*connection{"127.0.0.1:5600": conn}
	p.broken = map[uint64]destConn{}
	p.retry = map[string]chan struct{}{}
	p.Unlock()

	wg.Add(1)
	go p.changer(wg, ch, func(addr net.Addr, err error) {
		errs = append(errs, err)
	})

	ch <- addrUpdate{
		err: tErr,
	}
	ch <- addrUpdate{
		op:   addrUpdateOpDel,
		addr: "127.0.0.1:5600",
	}
	ch <- addrUpdate{
		op:   addrUpdateOpAdd,
		addr: "127.0.0.2:5600",
	}
	close(ch)

	wg.Wait()

	p.Lock()
	i2 := atomic.LoadUint64(c.autoID)
	assert.Equal(t, []error{tErr}, errs)
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	assert.Equal(t, map[uint64]destConn{
		i1: {c: conn},
		i2: {d: "127.0.0.2:5600"},
	}, p.broken)
	assert.Empty(t, p.retry)
	p.Unlock()
}

func TestProviderAddAddress(t *testing.T) {
	p := new(provider)
	p.wCnd = sync.NewCond(p)
	c := NewClient().(*client)
	p.c = c

	p.queue = []*connection{}
	p.healthy = map[string]*connection{}
	p.broken = map[uint64]destConn{}
	p.retry = map[string]chan struct{}{}

	p.addAddress("127.0.0.1:5600")
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	assert.Empty(t, p.broken)
	assert.Empty(t, p.retry)

	p.started = true
	p.addAddress("127.0.0.1:5600")
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	i := atomic.LoadUint64(c.autoID)
	assert.Equal(t, map[uint64]destConn{i: {d: "127.0.0.1:5600"}}, p.broken)
	assert.Empty(t, p.retry)

	conn := &connection{
		i: c.nextID(),
	}
	p.queue = []*connection{conn}
	p.healthy = map[string]*connection{"127.0.0.1:5600": conn}
	p.broken = map[uint64]destConn{}
	p.retry = map[string]chan struct{}{}
	p.addAddress("127.0.0.1:5600")
	assert.Equal(t, []*connection{conn}, p.queue)
	assert.Equal(t, map[string]*connection{"127.0.0.1:5600": conn}, p.healthy)
	assert.Empty(t, p.broken)
	assert.Empty(t, p.retry)

	p.queue = []*connection{}
	p.healthy = map[string]*connection{}
	p.broken = map[uint64]destConn{}
	p.retry = map[string]chan struct{}{"127.0.0.1:5600": nil}
	p.addAddress("127.0.0.1:5600")
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	assert.Empty(t, p.broken)
	assert.Equal(t, map[string]chan struct{}{"127.0.0.1:5600": nil}, p.retry)

	p.queue = []*connection{}
	p.healthy = map[string]*connection{}
	p.broken = map[uint64]destConn{conn.i: {d: "127.0.0.1:5600", c: conn}}
	p.retry = map[string]chan struct{}{}
	p.addAddress("127.0.0.1:5600")
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	assert.Equal(t, map[uint64]destConn{conn.i: {d: "127.0.0.1:5600", c: conn}}, p.broken)
	assert.Empty(t, p.retry)
}

func TestProviderDelAddress(t *testing.T) {
	p := new(provider)
	p.wCnd = sync.NewCond(p)
	c := NewClient().(*client)
	p.c = c

	p.queue = []*connection{}
	p.healthy = map[string]*connection{}
	p.broken = map[uint64]destConn{}
	p.retry = map[string]chan struct{}{}

	p.delAddress("127.0.0.1:5600")
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	assert.Empty(t, p.broken)
	assert.Empty(t, p.retry)

	p.started = true
	conn := &connection{
		i: c.nextID(),
	}
	p.queue = []*connection{conn}
	p.healthy = map[string]*connection{"127.0.0.1:5600": conn}
	p.broken = map[uint64]destConn{}
	p.retry = map[string]chan struct{}{}
	p.delAddress("127.0.0.1:5600")
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	assert.Equal(t, map[uint64]destConn{conn.i: {c: conn}}, p.broken)
	assert.Empty(t, p.retry)

	p.queue = []*connection{}
	p.healthy = map[string]*connection{}
	p.broken = map[uint64]destConn{}
	ch := make(chan struct{}, 1)
	p.retry = map[string]chan struct{}{"127.0.0.1:5600": ch}
	p.delAddress("127.0.0.1:5600")
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	assert.Empty(t, p.broken)
	assert.Empty(t, p.retry)
	select {
	default:
		assert.Fail(t, "retry channel hasn't been closed")
	case _, ok := <-ch:
		assert.False(t, ok)
	}

	p.queue = []*connection{}
	p.healthy = map[string]*connection{}
	p.broken = map[uint64]destConn{c.nextID(): {d: "127.0.0.1:5600"}}
	p.retry = map[string]chan struct{}{}
	p.delAddress("127.0.0.1:5600")
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	assert.Empty(t, p.broken)
	assert.Empty(t, p.retry)

	p.queue = []*connection{}
	p.healthy = map[string]*connection{}
	p.broken = map[uint64]destConn{conn.i: {d: "127.0.0.1:5600", c: conn}}
	p.retry = map[string]chan struct{}{}
	p.delAddress("127.0.0.1:5600")
	assert.Empty(t, p.queue)
	assert.Empty(t, p.healthy)
	assert.Equal(t, map[uint64]destConn{conn.i: {c: conn}}, p.broken)
	assert.Empty(t, p.retry)
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

type testProviderSyncDialerBrokenConn struct {
	ch  chan *testProviderBrokenConn
	err error
}

func newTestProviderSyncDialerBrokenConn(err error) testProviderSyncDialerBrokenConn {
	return testProviderSyncDialerBrokenConn{
		ch:  make(chan *testProviderBrokenConn),
		err: err,
	}
}

func (d testProviderSyncDialerBrokenConn) dial(a string) (net.Conn, error) {
	c := newTestProviderBrokenConn(a, d.err)
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

type testProviderBrokenConn struct {
	a      string
	err    error
	closed *uint32
	ch     chan struct{}
}

func newTestProviderBrokenConn(a string, err error) *testProviderBrokenConn {
	return &testProviderBrokenConn{
		a:      a,
		err:    err,
		closed: new(uint32),
		ch:     make(chan struct{}),
	}
}

func (c *testProviderBrokenConn) Close() error {
	if atomic.CompareAndSwapUint32(c.closed, 0, 1) {
		close(c.ch)
	}

	return c.err
}

func (c *testProviderBrokenConn) Write(b []byte) (int, error) {
	return len(b), nil
}

func (c *testProviderBrokenConn) Read(b []byte) (int, error) {
	<-c.ch
	return 0, io.EOF
}

func (c *testProviderBrokenConn) RemoteAddr() net.Addr {
	a, _ := net.ResolveTCPAddr("tcp", c.a)
	return a
}

func (c *testProviderBrokenConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c *testProviderBrokenConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c *testProviderBrokenConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c *testProviderBrokenConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }

type testProviderRadar struct {
	sync.Mutex

	ch chan addrUpdate
}

func (r *testProviderRadar) start(addrs []string) <-chan addrUpdate {
	r.Lock()
	defer r.Unlock()

	if r.ch != nil {
		return nil
	}

	r.ch = make(chan addrUpdate)
	return r.ch
}

func (r *testProviderRadar) stop() {
	r.Lock()
	defer r.Unlock()

	if r.ch != nil {
		close(r.ch)
		r.ch = nil
	}
}
