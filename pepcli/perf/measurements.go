package perf

import (
	"fmt"
	"sync"
	"time"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"
)

func measurement(c pep.Client, n, routineLimit int, rateLimit int64, noDump bool, reqs []*pb.Msg, maxResponseObligations uint32) ([]timing, error) {
	var pause time.Duration
	if rateLimit > 0 {
		pause = time.Second / time.Duration(rateLimit)
	}

	if noDump {
		if pause <= 0 && routineLimit > 0 {
			return parallelWithLimitNoDump(c, n, routineLimit, reqs, maxResponseObligations)
		}
	}

	if pause > 0 {
		if routineLimit > 0 {
			return parallelWithLimitAndPause(c, n, routineLimit, pause, reqs, maxResponseObligations)
		}

		if routineLimit < 0 {
			return parallelWithPause(c, n, pause, reqs, maxResponseObligations)
		}

		return sequentialWithPause(c, n, pause, reqs, maxResponseObligations)
	}

	if routineLimit > 0 {
		return parallelWithLimit(c, n, routineLimit, reqs, maxResponseObligations)
	}

	if routineLimit < 0 {
		return parallel(c, n, reqs, maxResponseObligations)
	}

	return sequential(c, n, reqs, maxResponseObligations)
}

func sequential(c pep.Client, n int, reqs []*pb.Msg, maxResponseObligations uint32) ([]timing, error) {
	out := make([]timing, n)

	var res pdp.Response
	obligation := make([]pdp.AttributeAssignment, maxResponseObligations)

	for i := 0; i < n; i++ {
		idx := i % len(reqs)

		res.Obligations = obligation

		out[i].setSend()
		err := c.Validate(reqs[idx], &res)
		if err != nil {
			return nil, fmt.Errorf("can't send request %d (%d): %s", idx, i, err)
		}
		out[i].setReceive()
	}

	return out, nil
}

func sequentialWithPause(c pep.Client, n int, p time.Duration, reqs []*pb.Msg, maxResponseObligations uint32) ([]timing, error) {
	out := make([]timing, n)

	var res pdp.Response
	obligation := make([]pdp.AttributeAssignment, maxResponseObligations)

	for i := 0; i < n; i++ {
		idx := i % len(reqs)

		res.Obligations = obligation

		out[i].setSend()
		err := c.Validate(reqs[idx], &res)
		if err != nil {
			return nil, fmt.Errorf("can't send request %d (%d): %s", idx, i, err)
		}
		out[i].setReceive()

		time.Sleep(p)
	}

	return out, nil
}

func parallel(c pep.Client, n int, reqs []*pb.Msg, maxResponseObligations uint32) ([]timing, error) {
	out := make([]timing, n)

	pool := makeObligationsPool(maxResponseObligations)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, req *pb.Msg) {
			obligation := pool.get()

			defer func() {
				pool.put(obligation)
				wg.Done()
			}()

			var res pdp.Response
			res.Obligations = obligation

			out[i].setSend()
			err := c.Validate(req, &res)
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

func parallelWithPause(c pep.Client, n int, p time.Duration, reqs []*pb.Msg, maxResponseObligations uint32) ([]timing, error) {
	out := make([]timing, n)

	pool := makeObligationsPool(maxResponseObligations)

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int, req *pb.Msg) {
			obligation := pool.get()

			defer func() {
				pool.put(obligation)
				wg.Done()
			}()

			var res pdp.Response
			res.Obligations = obligation

			out[i].setSend()
			err := c.Validate(req, &res)
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

func parallelWithLimit(c pep.Client, n, l int, reqs []*pb.Msg, maxResponseObligations uint32) ([]timing, error) {
	out := make([]timing, n)

	obligations := make(chan []pdp.AttributeAssignment, l)
	for i := 0; i < cap(obligations); i++ {
		obligations <- make([]pdp.AttributeAssignment, maxResponseObligations)
	}

	ch := make(chan int, l)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		ch <- 0

		wg.Add(1)
		go func(i int, req *pb.Msg) {
			obligation := <-obligations

			defer func() {
				obligations <- obligation
				wg.Done()
				<-ch
			}()

			var res pdp.Response
			res.Obligations = obligation

			out[i].setSend()
			err := c.Validate(req, &res)
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

func parallelWithLimitNoDump(c pep.Client, n, l int, reqs []*pb.Msg, maxResponseObligations uint32) ([]timing, error) {
	obligations := make(chan []pdp.AttributeAssignment, l)
	for i := 0; i < cap(obligations); i++ {
		obligations <- make([]pdp.AttributeAssignment, maxResponseObligations)
	}

	intervals := make([]time.Duration, l)
	counters := make([]int, l)

	ch := make(chan int, l)
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < n; i++ {
		ch <- 0

		wg.Add(1)
		go func(i int, req *pb.Msg) {
			obligation := <-obligations

			defer func() {
				obligations <- obligation
				wg.Done()
				<-ch
			}()

			var res pdp.Response
			res.Obligations = obligation

			start := time.Now()
			if err := c.Validate(req, &res); err == nil {
				intervals[i%l] += time.Now().Sub(start)
				counters[i%l]++
			}
		}(i, reqs[i%len(reqs)])
	}

	wg.Wait()

	var (
		tt    int64
		count int
	)

	for i, dt := range intervals {
		tt += dt.Nanoseconds()
		count += counters[i]
	}

	fmt.Printf("Rate: %.02f\nLatency: %s\n",
		float64(n)*1e9/float64(time.Now().Sub(start).Nanoseconds()),
		time.Duration(tt/int64(count)),
	)

	return nil, nil
}

func parallelWithLimitAndPause(c pep.Client, n, l int, p time.Duration, reqs []*pb.Msg, maxResponseObligations uint32) ([]timing, error) {
	out := make([]timing, n)

	obligations := make(chan []pdp.AttributeAssignment, l)
	for i := 0; i < cap(obligations); i++ {
		obligations <- make([]pdp.AttributeAssignment, maxResponseObligations)
	}

	ch := make(chan int, l)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		ch <- 0

		wg.Add(1)
		go func(i int, req *pb.Msg) {
			obligation := <-obligations

			defer func() {
				obligations <- obligation
				wg.Done()
				<-ch
			}()

			var res pdp.Response
			res.Obligations = obligation

			out[i].setSend()
			err := c.Validate(req, &res)
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
