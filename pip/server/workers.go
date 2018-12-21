package server

import "sync"

// ServiceHandler is a prototye for service handler function. The handler must write response and return the same buffer it got. It must not change buffer capacity.
type ServiceHandler func([]byte) []byte

func startWorkers(in chan []byte, n int, f ServiceHandler) chan []byte {
	out := make(chan []byte, n)

	wg := new(sync.WaitGroup)
	wg.Add(n)

	for i := 0; i < n; i++ {
		go worker(wg, in, out, f)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func worker(wg *sync.WaitGroup, in, out chan []byte, f ServiceHandler) {
	defer wg.Done()

	for msg := range in {
		out <- f(msg)
	}
}

func echo(b []byte) []byte {
	return b
}
