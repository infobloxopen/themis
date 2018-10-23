package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByteBufferPool(t *testing.T) {
	p := makeByteBufferPool(10)

	b := p.Get()
	assert.Equal(t, 10, len(b.b))

	p.Put(b)
}
