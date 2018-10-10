package client

import (
	"errors"
	"net"
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
		assert.Equal(t, c, conn.c)
		assert.Equal(t, n, conn.n)
		assert.NotZero(t, conn.r)
		assert.NotZero(t, conn.t)
		assert.NotZero(t, conn.p.idx)
	}
}

func TestConnectionGet(t *testing.T) {
	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	var sErr error
	go func() {
		sErr = s.Serve()
	}()
	defer func() {
		assert.NoError(t, sErr)
	}()

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
	s := server.NewServer()
	if !assert.NoError(t, s.Bind()) {
		assert.FailNow(t, "failed to bind server")
	}
	defer func() {
		assert.NoError(t, s.Stop())
	}()
	var sErr error
	go func() {
		sErr = s.Serve()
	}()
	defer func() {
		assert.NoError(t, sErr)
	}()

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

func TestIsConnectionClosed(t *testing.T) {
	err := errors.New("test")
	assert.False(t, isConnClosed(err))

	err = &net.OpError{
		Err: errors.New(netConnClosedMsg),
	}
	assert.True(t, isConnClosed(err))
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
