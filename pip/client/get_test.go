package client

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pip/server"
)

func TestClientGet(t *testing.T) {
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
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		v, err := c.Get()
		assert.Equal(t, pdp.UndefinedValue, v)
		assert.NoError(t, err)
	}
}

func TestClientGetErrNotConnected(t *testing.T) {
	c := NewClient()
	_, err := c.Get()
	assert.Equal(t, ErrNotConnected, err)
}

func TestClientGetMarshallingError(t *testing.T) {
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
	if err := c.Connect(); assert.NoError(t, err) {
		defer c.Close()

		_, err := c.Get(pdp.MakeExpressionAssignment("test", pdp.UndefinedValue))
		assert.Error(t, err)
	}
}
