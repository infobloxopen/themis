package client

import (
	"sync"
	"testing"
)

func TestTerminatorIncFirst(t *testing.T) {
	wg := new(sync.WaitGroup)

	inc := make(chan int)
	dec := make(chan int)

	wg.Add(1)
	go terminator(wg, inc, dec)

	inc <- 0
	inc <- 1
	inc <- 2
	close(inc)

	dec <- 0
	dec <- 1
	dec <- 2
	close(dec)

	wg.Wait()
}

func TestTerminatorDecFirst(t *testing.T) {
	wg := new(sync.WaitGroup)

	inc := make(chan int)
	dec := make(chan int)

	wg.Add(1)
	go terminator(wg, inc, dec)

	dec <- 0
	dec <- 1
	dec <- 2
	close(dec)

	inc <- 0
	inc <- 1
	inc <- 2
	close(inc)

	wg.Wait()
}
