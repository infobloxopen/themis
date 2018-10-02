package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	c := NewClient()
	assert.NotEmpty(t, c)
}

func TestNewClientWithTooSmallBuffer(t *testing.T) {
	c := NewClient(WithBufferSize(1024))
	if assert.NotEmpty(t, c) {
		if c, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, defBufSize, c.opts.bufSize)
		}
	}
}
