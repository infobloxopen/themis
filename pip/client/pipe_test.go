package client

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakePipe(t *testing.T) {
	p := makePipe()
	assert.NotZero(t, p)
}

func TestPipeClean(t *testing.T) {
	p := makePipe()

	p.putBytes(nil)
	assert.NotEmpty(t, p)

	p.clean()
	assert.Empty(t, p)

	p.clean()
	assert.Empty(t, p)
}

func TestPipeGet(t *testing.T) {
	p := makePipe()
	assert.Empty(t, p)

	p.putBytes(nil)

	b, err := p.get()
	assert.NoError(t, err)
	assert.Empty(t, b)
}

func TestPipePutBytes(t *testing.T) {
	p := makePipe()
	assert.Empty(t, p)

	p.putBytes([]byte{0xde, 0xc0, 0xad, 0xde})

	b, err := p.get()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b)
}

func TestPipePutError(t *testing.T) {
	p := makePipe()
	assert.Empty(t, p)

	tErr := errors.New("test")
	p.putError(tErr)

	b, err := p.get()
	assert.Equal(t, tErr, err)
	assert.Empty(t, b)
}
