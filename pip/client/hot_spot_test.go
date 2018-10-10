package client

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/themis/pip/server"
)

func TestHotSpotBalancerStart(t *testing.T) {
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
	b := new(hotSpotBalancer)

	if assert.NoError(t, b.start(c)) {
		defer b.stop()

		assert.NotZero(t, b.idx)
		assert.NotEmpty(t, b.conns)
	}
}

func TestHotSpotBalancerStartInvalidAddress(t *testing.T) {
	c := NewClient(
		WithAddress("localhost::5600"),
	).(*client)

	b := new(hotSpotBalancer)
	assert.Error(t, b.start(c))
}

func TestHotSpotBalancerStartWithConnectionError(t *testing.T) {
	c := NewClient().(*client)
	b := new(hotSpotBalancer)
	assert.Error(t, b.start(c))
}

func TestHotSpotBalancerGet(t *testing.T) {
	c := NewClient(
		WithMaxQueue(2),
		WithHotSpotBalancer(),
	).(*client)
	conn := c.b.get()
	assert.Zero(t, conn)

	b := c.b.(*hotSpotBalancer)
	b.idx = new(uint64)

	conn = c.b.get()
	assert.Zero(t, conn)

	first := c.newConnection(makeHSBTestConn())
	second := c.newConnection(makeHSBTestConn())
	b.conns = []*connection{first, second}

	conn = c.b.get()
	if assert.NotZero(t, conn) {
		assert.Equal(t, first, conn)
		assert.NotPanics(t, conn.g.Done)
	}

	conn.r <- request{}
	conn = c.b.get()
	if assert.NotZero(t, conn) {
		assert.Equal(t, first, conn)
		assert.NotPanics(t, conn.g.Done)
	}

	conn.r <- request{}
	conn = c.b.get()
	if assert.NotZero(t, conn) {
		assert.Equal(t, second, conn)
		assert.NotPanics(t, conn.g.Done)
	}
}

type hsbTestConn struct{}

func makeHSBTestConn() hsbTestConn                       { return hsbTestConn{} }
func (c hsbTestConn) Close() error                       { panic("not implemented") }
func (c hsbTestConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c hsbTestConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c hsbTestConn) Write(b []byte) (int, error)        { panic("not implemented") }
func (c hsbTestConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c hsbTestConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c hsbTestConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c hsbTestConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
