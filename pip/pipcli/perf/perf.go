package perf

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pip/pipcli/global"
)

var (
	workers int

	errUndefinedResponse = errors.New("undefined response")
)

func command(conf *global.Config) error {
	n := conf.N
	if n <= 0 {
		n = len(conf.Requests)
	}

	p := new(int64)
	*p = -1

	out := make([]timing, n)

	wg := new(sync.WaitGroup)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			for {
				j := int(atomic.AddInt64(p, 1))
				if j >= n {
					return
				}

				out[j].setSend()
				r := conf.Requests[j%len(conf.Requests)]
				v, err := conf.Client.Get(r.Path, r.Args)
				if err != nil {
					out[j].setError(err)
					return
				}

				if v == pdp.UndefinedValue {
					out[j].setError(errUndefinedResponse)
					return
				}
				out[j].setReceive()
			}
		}(i)
	}

	wg.Wait()

	for _, t := range out {
		if t.e != nil {
			return t.e
		}
	}

	return dump(out, "")
}
