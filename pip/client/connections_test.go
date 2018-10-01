package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientConnect(t *testing.T) {
	c := NewClient()

	if assert.NoError(t, c.Connect()) {
		assert.Equal(t, ErrorConnected, c.Connect())
	}
}

func TestClientClose(t *testing.T) {
	c := NewClient()

	if assert.NoError(t, c.Connect()) {
		c.Close()
	}
}
