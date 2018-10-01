package client

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/themis/pdp"
)

func TestClientGet(t *testing.T) {
	c := NewClient()

	v, err := c.Get()
	assert.Equal(t, pdp.UndefinedValue, v)
	assert.Equal(t, errNotImplemented, err)
}
