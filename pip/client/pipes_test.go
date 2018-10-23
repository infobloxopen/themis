package client

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMakePipes(t *testing.T) {
	p := makePipes(3, defTimeout.Nanoseconds())
	assert.Equal(t, 3, len(p.idx))
	assert.Equal(t, len(p.idx), cap(p.idx))
	assert.Equal(t, len(p.idx), len(p.p))
}

func TestPipesAllocFree(t *testing.T) {
	p := makePipes(3, defTimeout.Nanoseconds())

	seq := []int{}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	wg3 := new(sync.WaitGroup)
	wg3.Add(1)
	go func() {
		defer wg.Done()

		for i := 0; i < cap(p.idx)+1; i++ {
			if i == 3 {
				wg3.Done()
			}

			idx, _ := p.alloc()
			seq = append(seq, idx)
		}
	}()

	wg3.Wait()

	assert.ElementsMatch(t, []int{0, 1, 2}, seq)
	p.free(2)

	wg.Wait()

	assert.ElementsMatch(t, []int{0, 1, 2, 2}, seq)
}

func TestPipesPutBytes(t *testing.T) {
	ps := makePipes(3, defTimeout.Nanoseconds())

	i, p := ps.alloc()
	b := &byteBuffer{
		b: []byte{0xde, 0xc0, 0xad, 0xde},
	}
	ps.putBytes(i, b)

	b, err := p.get()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b.b)
}

func TestPipesPutError(t *testing.T) {
	ps := makePipes(3, defTimeout.Nanoseconds())

	i, p := ps.alloc()
	tErr := errors.New("test")
	ps.putError(i, tErr)

	b, err := p.get()
	assert.Equal(t, tErr, err)
	assert.Empty(t, b)
}

func TestPipesCheck(t *testing.T) {
	ps := makePipes(1, defTimeout.Nanoseconds())

	i, p := ps.alloc()
	defer ps.free(i)

	assert.False(t, ps.check(time.Unix(0, *p.t).Add(defTimeout/2)))
	assert.True(t, ps.check(time.Unix(0, *p.t).Add(2*defTimeout)))
}

func TestPipesFlush(t *testing.T) {
	ps := makePipes(1, defTimeout.Nanoseconds())

	i, p := ps.alloc()
	var err error
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer ps.free(i)

		_, err = p.get()
	}()

	ps.flush()

	wg.Wait()
	assert.Equal(t, errReaderBroken, err)
}
