package client

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/themis/pip/server"
)

func TestNewConnection(t *testing.T) {
	c := NewClient().(*client)

	n := makeCTestConn()
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
	conn := c.newConnection(n)
	conn.start()
	defer conn.close()

	b := c.pool.Get()
	defer c.pool.Put(b)

	b = append(b[:0], 0xef, 0xbe, 0xad, 0xde)
	b, err = conn.get(b)
	if b != nil {
		c.pool.Put(b[:cap(b)])
	}

	assert.NoError(t, err)
	assert.Equal(t, []byte{0xef, 0xbe, 0xad, 0xde}, b)
}

type cTestConn struct{}

func makeCTestConn() cTestConn                         { return cTestConn{} }
func (c cTestConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c cTestConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c cTestConn) Close() error                       { panic("not implemented") }
func (c cTestConn) Write(b []byte) (n int, err error)  { panic("not implemented") }
func (c cTestConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c cTestConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c cTestConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c cTestConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
