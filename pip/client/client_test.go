package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	c := NewClient()
	assert.NotEmpty(t, c)
}

func TestNewClientWithNetwork(t *testing.T) {
	c := NewClient(WithNetwork("unix"))
	if assert.NotEmpty(t, c) {
		if c, ok := c.(*client); assert.True(t, ok) {
			assert.Equal(t, "unix", c.opts.net)
		}
	}
}
