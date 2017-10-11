package perf

import (
	"fmt"
	"sync"
	"time"

	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"
)

func measurement(c pep.Client, n, routineLimit int, rateLimit int64, reqs []pdp.Request) ([]timing, error) {
	var pause time.Duration
	if rateLimit > 0 {
		pause = time.Second / time.Duration(rateLimit)
	}

	if pause > 0 {
		if routineLimit > 0 {
			return parallelWithLimitAndPause(c, n, routineLimit, pause, reqs)
		}

		if routineLimit < 0 {
			return parallelWithPause(c, n, pause, reqs)
		}

		return sequentialWithPause(c, n, pause, reqs)
	}

	if routineLimit > 0 {
		return parallelWithLimit(c, n, routineLimit, reqs)
	}

	if routineLimit < 0 {
		return parallel(c, n, reqs)
	}

	return sequential(c, n, reqs)
}

func sequential(c pep.Client, n int, reqs []pdp.Request) ([]timing, error) {
	out := make([]timing, n)

	for i := 0; i < n; i++ {
		idx := i % len(reqs)

		out[i].setSend()
		_, err := c.Validate(&reqs[idx])
		if err != nil {
			return nil, fmt.Errorf("can't send request %d (%d): %s", idx, i, err)
		}
		out[i].setReceive()
	}

	return out, nil
}

func sequentialWithPause(c pep.Client, n int, p time.Duration, reqs []pdp.Request) ([]timing, error) {
	out := make([]timing, n)

	for i := 0; i < n; i++ {
		idx := i % len(reqs)

		out[i].setSend()
		_, err := c.Validate(&reqs[idx])
		if err != nil {
			return nil, fmt.Errorf("can't send request %d (%d): %s", idx, i, err)
		}
		out[i].setReceive()

		time.Sleep(p)
	}

	return out, nil
}

func parallel(c pep.Client, n int, reqs []pdp.Request) ([]timing, error) {
	out := make([]timing, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, req pdp.Request) {
			defer wg.Done()

			out[i].setSend()
			_, err := c.Validate(&req)
			if err != nil {
				out[i].setError(err)
			} else {
				out[i].setReceive()
			}
		}(i, reqs[i%len(reqs)])
	}

	wg.Wait()

	return out, nil
}

func parallelWithPause(c pep.Client, n int, p time.Duration, reqs []pdp.Request) ([]timing, error) {
	out := make([]timing, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, req pdp.Request) {
			defer wg.Done()

			out[i].setSend()
			_, err := c.Validate(&req)
			if err != nil {
				out[i].setError(err)
			} else {
				out[i].setReceive()
			}
		}(i, reqs[i%len(reqs)])

		time.Sleep(p)
	}

	wg.Wait()

	return out, nil
}

func parallelWithLimit(c pep.Client, n, l int, reqs []pdp.Request) ([]timing, error) {
	out := make([]timing, n)

	ch := make(chan int, l)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		ch <- 0

		wg.Add(1)
		go func(i int, req pdp.Request) {
			defer func() {
				wg.Done()
				<-ch
			}()

			out[i].setSend()
			_, err := c.Validate(&req)
			if err != nil {
				out[i].setError(err)
			} else {
				out[i].setReceive()
			}
		}(i, reqs[i%len(reqs)])
	}

	wg.Wait()

	return out, nil
}

func parallelWithLimitAndPause(c pep.Client, n, l int, p time.Duration, reqs []pdp.Request) ([]timing, error) {
	out := make([]timing, n)

	ch := make(chan int, l)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		ch <- 0

		wg.Add(1)
		go func(i int, req pdp.Request) {
			defer func() {
				wg.Done()
				<-ch
			}()

			out[i].setSend()
			_, err := c.Validate(&req)
			if err != nil {
				out[i].setError(err)
			} else {
				out[i].setReceive()
			}
		}(i, reqs[i%len(reqs)])

		time.Sleep(p)
	}

	wg.Wait()

	return out, nil
}
