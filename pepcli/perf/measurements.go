package perf

import (
	"fmt"
	"sync"
	"time"

	pb "github.com/infobloxopen/themis/pdp-service"
)

type validator func(in, out interface{}) error

func measurement(v validator, n, routineLimit int, rateLimit int64, reqs []pb.Request) ([]timing, error) {
	var pause time.Duration
	if rateLimit > 0 {
		pause = time.Second / time.Duration(rateLimit)
	}

	if pause > 0 {
		if routineLimit > 0 {
			return parallelWithLimitAndPause(v, n, routineLimit, pause, reqs)
		}

		if routineLimit < 0 {
			return parallelWithPause(v, n, pause, reqs)
		}

		return sequentialWithPause(v, n, pause, reqs)
	}

	if routineLimit > 0 {
		return parallelWithLimit(v, n, routineLimit, reqs)
	}

	if routineLimit < 0 {
		return parallel(v, n, reqs)
	}

	return sequential(v, n, reqs)
}

func sequential(v validator, n int, reqs []pb.Request) ([]timing, error) {
	out := make([]timing, n)
	res := &pb.Response{}

	for i := 0; i < n; i++ {
		idx := i % len(reqs)

		out[i].setSend()
		err := v(reqs[idx], res)
		if err != nil {
			return nil, fmt.Errorf("can't send request %d (%d): %s", idx, i, err)
		}
		out[i].setReceive()
	}

	return out, nil
}

func sequentialWithPause(v validator, n int, p time.Duration, reqs []pb.Request) ([]timing, error) {
	out := make([]timing, n)
	res := &pb.Response{}

	for i := 0; i < n; i++ {
		idx := i % len(reqs)

		out[i].setSend()
		err := v(reqs[idx], res)
		if err != nil {
			return nil, fmt.Errorf("can't send request %d (%d): %s", idx, i, err)
		}
		out[i].setReceive()

		time.Sleep(p)
	}

	return out, nil
}

func parallel(v validator, n int, reqs []pb.Request) ([]timing, error) {
	out := make([]timing, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, req pb.Request) {
			defer wg.Done()

			res := &pb.Response{}

			out[i].setSend()
			err := v(req, res)
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

func parallelWithPause(v validator, n int, p time.Duration, reqs []pb.Request) ([]timing, error) {
	out := make([]timing, n)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, req pb.Request) {
			defer wg.Done()

			res := &pb.Response{}

			out[i].setSend()
			err := v(req, res)
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

func parallelWithLimit(v validator, n, l int, reqs []pb.Request) ([]timing, error) {
	out := make([]timing, n)

	ch := make(chan int, l)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		ch <- 0

		wg.Add(1)
		go func(i int, req pb.Request) {
			defer func() {
				wg.Done()
				<-ch
			}()

			res := &pb.Response{}

			out[i].setSend()
			err := v(req, res)
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

func parallelWithLimitAndPause(v validator, n, l int, p time.Duration, reqs []pb.Request) ([]timing, error) {
	out := make([]timing, n)

	ch := make(chan int, l)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		ch <- 0

		wg.Add(1)
		go func(i int, req pb.Request) {
			defer func() {
				wg.Done()
				<-ch
			}()

			res := &pb.Response{}

			out[i].setSend()
			err := v(req, res)
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
