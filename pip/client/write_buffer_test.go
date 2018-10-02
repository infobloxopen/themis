package client

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWriteBuffer(t *testing.T) {
	p := makePipes(1)

	w := newWriteBuffer(16, p)
	if assert.NotZero(t, w) {
		assert.Equal(t, 16, cap(w.out))
		assert.Empty(t, w.out)
		assert.NotZero(t, w.idx)
		assert.Equal(t, p, w.p)
	}
}

func TestWriteBufferIsEmpty(t *testing.T) {
	w := newWriteBuffer(16, makePipes(1))

	assert.True(t, w.isEmpty())

	w.put(request{
		i: 0,
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	})
	assert.False(t, w.isEmpty())
}

func TestWriteBufferRem(t *testing.T) {
	w := newWriteBuffer(16, makePipes(1))

	assert.Equal(t, 16, w.rem())

	w.put(request{
		i: 0,
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	})
	assert.Equal(t, 4, w.rem())
}

func TestWriteBufferPut(t *testing.T) {
	wg := new(sync.WaitGroup)
	p := makePipes(2)

	w := newWriteBuffer(16, p)

	i1, p1 := p.alloc()
	wg.Add(1)
	err1 := errors.New("test1")
	go func() {
		defer wg.Done()
		defer p.free(i1)

		_, err1 = p1.get()
	}()

	w.put(request{
		i: i1,
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	})

	i2, p2 := p.alloc()
	wg.Add(1)
	err2 := errors.New("test2")
	go func() {
		defer wg.Done()
		defer p.free(i2)

		_, err2 = p2.get()
	}()

	w.put(request{
		i: i2,
		b: []byte{0xde, 0xc0, 0xad, 0xde, 0xef, 0xeb, 0, 0},
	})

	wg.Wait()
	assert.NoError(t, err1)
	assert.NoError(t, err2)
}

func TestWriteBufferFlush(t *testing.T) {
	wg := new(sync.WaitGroup)
	ps := makePipes(1)

	w := newWriteBuffer(16, ps)

	i, p := ps.alloc()
	wg.Add(1)
	err := errors.New("test")
	go func() {
		defer wg.Done()
		defer ps.free(i)

		_, err = p.get()
	}()

	w.put(request{
		i: i,
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	})
	assert.Empty(t, p)

	w.flush()

	wg.Wait()
	assert.NoError(t, err)
}
