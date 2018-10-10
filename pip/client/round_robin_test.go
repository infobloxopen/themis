package client

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/themis/pip/server"
)

func TestRoundRobinBalancerStart(t *testing.T) {
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
	b := new(roundRobinBalancer)

	if assert.NoError(t, b.start(c)) {
		defer b.stop()

		assert.NotZero(t, b.idx)
		assert.NotEmpty(t, b.conns)
	}
}

func TestRoundRobinBalancerStartInvalidAddress(t *testing.T) {
	c := NewClient(
		WithAddress("localhost::5600"),
	).(*client)

	b := new(roundRobinBalancer)
	assert.Error(t, b.start(c))
}

func TestRoundRobinBalancerStartWithConnectionError(t *testing.T) {
	c := NewClient().(*client)
	b := new(roundRobinBalancer)
	assert.Error(t, b.start(c))
}

func TestRoundRobinBalancerGet(t *testing.T) {
	c := NewClient(
		WithRoundRobinBalancer(),
	).(*client)
	conn := c.b.get()
	assert.Zero(t, conn)

	b := c.b.(*roundRobinBalancer)
	b.idx = new(uint64)

	conn = c.b.get()
	assert.Zero(t, conn)

	first := c.newConnection(makeRRBTestConn())
	second := c.newConnection(makeRRBTestConn())
	b.conns = []*connection{first, second}

	conn = c.b.get()
	if assert.NotZero(t, conn) {
		assert.Equal(t, first, conn)
		assert.NotPanics(t, conn.g.Done)
	}

	conn = c.b.get()
	if assert.NotZero(t, conn) {
		assert.Equal(t, second, conn)
		assert.NotPanics(t, conn.g.Done)
	}
}

type rrbTestConn struct{}

func makeRRBTestConn() rrbTestConn                       { return rrbTestConn{} }
func (c rrbTestConn) Close() error                       { panic("not implemented") }
func (c rrbTestConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c rrbTestConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c rrbTestConn) Write(b []byte) (int, error)        { panic("not implemented") }
func (c rrbTestConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c rrbTestConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c rrbTestConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c rrbTestConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
