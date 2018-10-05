package client

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTerminator(t *testing.T) {
	wg := new(sync.WaitGroup)
	c := NewClient().(*client)

	p := makePipes(defMaxQueue, defTimeout.Nanoseconds())
	dt := make(chan struct{})

	nc := newTestTerminatorConn()

	wg.Add(1)
	go c.terminator(wg, nc, p, dt)

	close(dt)
	wg.Wait()

	assert.False(t, nc.c)
}

func TestTerminatorTimeout(t *testing.T) {
	wg := new(sync.WaitGroup)
	ch := make(chan time.Time)
	c := NewClient(withTestTermFlushChannel(ch)).(*client)

	ps := makePipes(defMaxQueue, defTimeout.Nanoseconds())
	dt := make(chan struct{})

	nc := newTestTerminatorConn()

	wg.Add(1)
	go c.terminator(wg, nc, ps, dt)

	i1, p1 := ps.alloc()
	var err1 error
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ps.free(i1)

		_, err1 = p1.get()
	}()

	i2, p2 := ps.alloc()
	var err2 error
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ps.free(i2)

		_, err2 = p2.get()
	}()

	next := time.Unix(0, *p1.t).Add(2 * defTimeout)
	*p2.t = next.UnixNano()

	ch <- next

	close(dt)
	wg.Wait()

	assert.True(t, nc.c)
	assert.Equal(t, errTimeout, err1)
	assert.Equal(t, errReaderBroken, err2)
}

type testTerminatorConn struct {
	c bool
}

func newTestTerminatorConn() *testTerminatorConn {
	return new(testTerminatorConn)
}

func (c *testTerminatorConn) Close() error {
	c.c = true
	return nil
}

func (c *testTerminatorConn) Write(b []byte) (int, error)        { panic("not implemented") }
func (c *testTerminatorConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c *testTerminatorConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c *testTerminatorConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c *testTerminatorConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c *testTerminatorConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c *testTerminatorConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
