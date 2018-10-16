package client

import (
	"errors"
	"math"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/themis/pip/server"
)

func TestNewConnection(t *testing.T) {
	c := NewClient().(*client)

	n := makeCTestConn(nil)
	conn := c.newConnection(n)
	if assert.NotZero(t, conn) {
		assert.NotZero(t, conn.i)
		assert.Equal(t, c, conn.c)
		assert.Equal(t, n, conn.n)
		assert.NotZero(t, conn.r)
		assert.NotZero(t, conn.t)
		assert.NotZero(t, conn.p.idx)
	}
}

func TestNewConnectionOverflow(t *testing.T) {
	c := NewClient().(*client)
	*c.autoID = math.MaxUint64

	conn := c.newConnection(nil)
	assert.Zero(t, conn)
}

func TestConnectionGet(t *testing.T) {
	s := newTestServerForConn(t)
	defer s.stop(t)

	c := NewClient().(*client)

	n, err := net.Dial(c.opts.net, c.opts.addr)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to connect to test server")
	}

	conn := c.newConnection(n)
	conn.start()
	defer conn.close()

	b := c.pool.Get()
	defer func() {
		if b != nil {
			c.pool.Put(b)
		}
	}()

	b = append(b[:0], 0xef, 0xbe, 0xad, 0xde)
	b, err = conn.get(b)

	assert.NoError(t, err)
	assert.Equal(t, []byte{0xef, 0xbe, 0xad, 0xde}, b)
}

func TestConnectionIsFull(t *testing.T) {
	c := NewClient(
		WithMaxQueue(2),
	).(*client)

	n := makeCTestConn(nil)
	conn := c.newConnection(n)
	assert.False(t, conn.isFull())

	conn.r <- request{}
	assert.False(t, conn.isFull())

	conn.r <- request{}
	assert.True(t, conn.isFull())

	<-conn.r
	assert.False(t, conn.isFull())
}

func TestConnectionClose(t *testing.T) {
	s := newTestServerForConn(t)
	defer s.stop(t)

	var cErr error
	c := NewClient(WithConnErrHandler(func(a net.Addr, err error) {
		cErr = err
	})).(*client)

	n, err := net.Dial(c.opts.net, c.opts.addr)
	if !assert.NoError(t, err) {
		assert.FailNow(t, "failed to connect to test server")
	}

	conn := c.newConnection(n)
	conn.start()
	conn.close()
	assert.NoError(t, cErr)
}

func TestConnectionCloseNet(t *testing.T) {
	var cErr error
	c := NewClient(WithConnErrHandler(func(a net.Addr, err error) {
		cErr = err
	})).(*client)

	tErr := errors.New("test")
	conn := c.newConnection(makeCTestConn(tErr))
	conn.closeNet()
	assert.Equal(t, tErr, cErr)
}

func TestIsConnectionRefused(t *testing.T) {
	err := errors.New("test")
	assert.False(t, isConnRefused(err))

	err = &net.OpError{
		Err: errors.New("test"),
	}
	assert.False(t, isConnRefused(err))

	err = &net.OpError{
		Err: os.NewSyscallError("test", errors.New(netConnRefusedMsg)),
	}
	assert.True(t, isConnRefused(err))
}

func TestIsConnectionClosed(t *testing.T) {
	err := errors.New("test")
	assert.False(t, isConnClosed(err))

	err = &net.OpError{
		Err: errors.New(netConnClosedMsg),
	}
	assert.True(t, isConnClosed(err))
}

func TestIsConnectionTimeout(t *testing.T) {
	err := errors.New("test")
	assert.False(t, isConnTimeout(err))

	err = &net.OpError{
		Err: cTestTimeoutError{},
	}
	assert.True(t, isConnTimeout(err))
}

type testServerForConn struct {
	sync.WaitGroup
	s   *server.Server
	err error
}

func newTestServerForConn(t *testing.T) *testServerForConn {
	s := new(testServerForConn)

	s.s = server.NewServer()
	if !assert.NoError(t, s.s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}

	s.Add(1)
	go func() {
		defer s.Done()

		s.err = s.s.Serve()
	}()

	return s
}

func (s *testServerForConn) stop(t *testing.T) {
	if assert.NoError(t, s.s.Stop()) {
		s.Wait()

		if s.err != server.ErrNotBound {
			assert.NoError(t, s.err)
		}
	}
}

type cTestConn struct {
	err error
}

func makeCTestConn(err error) cTestConn {
	return cTestConn{
		err: err,
	}
}

func (c cTestConn) Close() error {
	return c.err
}

func (c cTestConn) RemoteAddr() net.Addr {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	return addr
}

func (c cTestConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c cTestConn) Write(b []byte) (int, error)        { panic("not implemented") }
func (c cTestConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c cTestConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c cTestConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c cTestConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }

type cTestTimeoutError struct{}

func (err cTestTimeoutError) Error() string {
	return "test timeout"
}

func (err cTestTimeoutError) Timeout() bool {
	return true
}
