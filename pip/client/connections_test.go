package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientConnect(t *testing.T) {
	c := NewClient()

	if assert.NoError(t, c.Connect()) {
		assert.Equal(t, ErrConnected, c.Connect())
	}
}

func TestClientClose(t *testing.T) {
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
