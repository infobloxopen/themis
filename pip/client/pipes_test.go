package client

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakePipes(t *testing.T) {
	p := makePipes(3)
	assert.Equal(t, 3, len(p.idx))
	assert.Equal(t, len(p.idx), cap(p.idx))
	assert.Equal(t, len(p.idx), len(p.p))
}

func TestPipesAllocFree(t *testing.T) {
	p := makePipes(3)

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
	ps := makePipes(3)

	i, p := ps.alloc()
	ps.putBytes(i, []byte{0xde, 0xc0, 0xad, 0xde})

	b, err := p.get()
	assert.NoError(t, err)
	assert.Equal(t, []byte{0xde, 0xc0, 0xad, 0xde}, b)
}

func TestPipesPutError(t *testing.T) {
	ps := makePipes(3)

	i, p := ps.alloc()
	tErr := errors.New("test")
	ps.putError(i, tErr)

	b, err := p.get()
	assert.Equal(t, tErr, err)
	assert.Empty(t, b)
}
