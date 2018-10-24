package client

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTerminator(t *testing.T) {
	c := NewClient().(*client)

	n := newTestTerminatorConn()
	conn := c.newConnection(n)

	conn.w.Add(1)
	go conn.terminator()

	close(conn.t)
	conn.w.Wait()

	assert.False(t, n.c)
}

func TestTerminatorTimeout(t *testing.T) {
	ch := make(chan time.Time)
	c := NewClient().(*client)
	c.opts.termFlushCh = ch

	n := newTestTerminatorConn()
	conn := c.newConnection(n)

	conn.w.Add(1)
	go conn.terminator()

	i1, p1 := conn.p.alloc()
	var err1 error
	conn.w.Add(1)
	go func() {
		defer conn.w.Done()
		defer conn.p.free(i1)

		_, err1 = p1.get()
	}()

	i2, p2 := conn.p.alloc()
	var err2 error
	conn.w.Add(1)
	go func() {
		defer conn.w.Done()
		defer conn.p.free(i2)

		_, err2 = p2.get()
	}()

	next := time.Unix(0, *p1.t).Add(2 * defTimeout)
	*p2.t = next.UnixNano()

	ch <- next

	close(conn.t)
	conn.w.Wait()

	assert.True(t, n.c)
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
