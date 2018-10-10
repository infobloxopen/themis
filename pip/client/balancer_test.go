package client

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/themis/pip/server"
)

func TestNewBalancer(t *testing.T) {
	b := newBalancer(defNet, balancerTypeSimple)
	assert.IsType(t, new(simpleBalancer), b)

	b = newBalancer(defNet, balancerTypeRoundRobin)
	assert.IsType(t, new(roundRobinBalancer), b)

	b = newBalancer(unixNet, balancerTypeRoundRobin)
	assert.IsType(t, new(simpleBalancer), b)

	b = newBalancer(defNet, balancerTypeHotSpot)
	assert.IsType(t, new(hotSpotBalancer), b)

	b = newBalancer(unixNet, balancerTypeHotSpot)
	assert.IsType(t, new(simpleBalancer), b)
}

func TestBalancerStart(t *testing.T) {
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
	b := new(simpleBalancer)

	if assert.NoError(t, b.start(c)) {
		defer b.stop()

		assert.NotZero(t, b.c)
	}
}

func TestBalancerStartError(t *testing.T) {
	c := NewClient(
		WithNetwork("unix"),
		WithAddress("/dev/null"),
	).(*client)

	b := new(simpleBalancer)
	err := b.start(c)
	if assert.Error(t, err) {
		assert.IsType(t, new(net.OpError), err)
	}
}

func TestBalancerGet(t *testing.T) {
	c := NewClient().(*client)
	conn := c.b.get()
	assert.Zero(t, conn)

	b := c.b.(*simpleBalancer)
	b.c = c.newConnection(makeBTestConn())
	conn = c.b.get()
	if assert.NotZero(t, conn) {
		assert.NotPanics(t, conn.g.Done)
	}
}

type bTestConn struct{}

func makeBTestConn() bTestConn                         { return bTestConn{} }
func (c bTestConn) Close() error                       { panic("not implemented") }
func (c bTestConn) RemoteAddr() net.Addr               { panic("not implemented") }
func (c bTestConn) Read(b []byte) (int, error)         { panic("not implemented") }
func (c bTestConn) Write(b []byte) (int, error)        { panic("not implemented") }
func (c bTestConn) LocalAddr() net.Addr                { panic("not implemented") }
func (c bTestConn) SetDeadline(t time.Time) error      { panic("not implemented") }
func (c bTestConn) SetReadDeadline(t time.Time) error  { panic("not implemented") }
func (c bTestConn) SetWriteDeadline(t time.Time) error { panic("not implemented") }
