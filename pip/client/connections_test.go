package client

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/themis/pip/server"
)

func TestClientConnect(t *testing.T) {
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

	c := NewClient()

	if assert.NoError(t, c.Connect()) {
		assert.Equal(t, ErrConnected, c.Connect())
	}
}

func TestClientConnectNoServer(t *testing.T) {
	c := NewClient()

	err := c.Connect()
	if assert.Error(t, err) {
		assert.IsType(t, (*net.OpError)(nil), err)
	}
}

func TestClientClose(t *testing.T) {
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

	c := NewClient()

	if assert.NoError(t, c.Connect()) {
		c.Close()
		if cc, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, pipClientIdle, *cc.state)
		}

		c.Close()
		if cc, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, pipClientIdle, *cc.state)
		}
	}
}
