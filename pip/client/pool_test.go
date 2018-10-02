package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBytePool(t *testing.T) {
	p := makeBytePool(10, false)

	b := p.Get()
	assert.Equal(t, 10, len(b))

	p.Put(b)
}

func TestDummyBytePool(t *testing.T) {
	p := makeBytePool(10, true)

	b := p.Get()
	assert.Equal(t, 10, len(b))

	p.Put(b)
}
