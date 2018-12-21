package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakePool(t *testing.T) {
	p := makePool(3, 8)
	assert.Equal(t, 3, len(p.ch))
	assert.Equal(t, cap(p.ch), len(p.ch))
}

func TestPoolGet(t *testing.T) {
	p := makePool(3, 8)
	b := p.get()
	assert.Equal(t, 0, len(b))
	assert.Equal(t, 8, cap(b))
}

func TestPoolPut(t *testing.T) {
	p := makePool(3, 8)

	var b [3][]byte
	b[0] = p.get()
	b[1] = p.get()
	b[2] = p.get()

	b[0] = append(b[0], 1, 2, 3, 4)
	b[1] = append(b[1], 5, 6, 7, 8, 9)
	b[2] = append(b[2], 10, 11, 12, 13, 14, 15)

	p.put(b[0])
	p.put(b[1])
	p.put(b[2])

	b[0] = p.get()
	assert.Equal(t, 0, len(b[0]))
	assert.Equal(t, 8, cap(b[0]))

	b[1] = p.get()
	assert.Equal(t, 0, len(b[1]))
	assert.Equal(t, 8, cap(b[1]))

	b[2] = p.get()
	assert.Equal(t, 0, len(b[2]))
	assert.Equal(t, 8, cap(b[2]))
}
