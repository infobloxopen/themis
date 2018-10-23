package client

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMakePipe(t *testing.T) {
	p := makePipe()
	if assert.NotZero(t, p.t) {
		assert.Equal(t, int64(math.MinInt64), *p.t)
	}
	assert.NotZero(t, p.ch)
}

func TestPipeClean(t *testing.T) {
	p := makePipe()

	p.putBytes(nil)
	assert.NotEmpty(t, p.ch)
	assert.Equal(t, int64(math.MinInt64), *p.t)

	n := new(int64)
	*n = time.Now().UnixNano()

	p.clean(n)
	assert.Empty(t, p.ch)
	assert.Equal(t, *n, *p.t)

	p.clean(n)
	assert.Empty(t, p.ch)
}

func TestPipeGet(t *testing.T) {
	p := makePipe()
	assert.Empty(t, p.ch)

	p.putBytes(nil)

	b, err := p.get()
	assert.NoError(t, err)
	assert.Empty(t, b)
}

func TestPipePutBytes(t *testing.T) {
	p := makePipe()
	assert.Empty(t, p.ch)

	b := &byteBuffer{
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	}
	assert.True(t, p.putBytes(b))

	b, err := p.get()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b.b)
}

func TestPipePutBytesDuplicate(t *testing.T) {
	p := makePipe()
	assert.Empty(t, p.ch)

	b1 := &byteBuffer{
		b: []byte{0x01, 0x02, 0x03, 0x04},
	}
	assert.True(t, p.putBytes(b1))

	b2 := &byteBuffer{
		b: []byte{0x05, 0x06, 0x07, 0x08},
	}
	assert.False(t, p.putBytes(b2))

	b, err := p.get()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04}, b.b)
}

func TestPipePutError(t *testing.T) {
	p := makePipe()
	assert.Empty(t, p.ch)

	tErr := errors.New("test")
	p.putError(tErr)

	b, err := p.get()
	assert.Equal(t, tErr, err)
	assert.Empty(t, b)
}
