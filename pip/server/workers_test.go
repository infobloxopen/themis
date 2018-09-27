package server

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStartWorkers(t *testing.T) {
	in := make(chan []byte, 8)
	out := startWorkers(in, 2, echo)

	go func() {
		defer close(in)

		for i := 0; i < cap(in); i++ {
			in <- make([]byte, i+1)
		}
	}()

	msgs := [][]byte{}
	for msg := range out {
		msgs = append(msgs, msg)
	}

	assert.ElementsMatch(t, [][]byte{
		{0},
		{0, 0},
		{0, 0, 0},
		{0, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 0},
	}, msgs)
}

func TestWorker(t *testing.T) {
	wg := new(sync.WaitGroup)

	in := make(chan []byte, 3)
	for len(in) < cap(in) {
		in <- nil
	}
	close(in)

	out := make(chan []byte, 3)

	wg.Add(1)
	worker(wg, in, out, echo)
	wg.Wait()

	assert.Equal(t, cap(out), len(out))
}

func TestEcho(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	b := echo(a)
	assert.Equal(t, a, b)
}
