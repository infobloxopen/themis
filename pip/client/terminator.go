package client

import "sync"

func terminator(wg *sync.WaitGroup, inc, dec chan int) {
	defer wg.Done()

	for {
		select {
		case _, ok := <-inc:
			if !ok {
				for range dec {
				}

				return
			}

		case _, ok := <-dec:
			if !ok {
				for range inc {
				}

				return
			}
		}
	}
}
