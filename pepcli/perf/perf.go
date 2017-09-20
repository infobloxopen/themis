package perf

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"

	"github.com/infobloxopen/themis/pepcli/requests"
)

const (
	Name        = "perf"
	Description = "measures performance of evaluation given requests on PDP server"
)

func Exec(addr, in, out string, n int, v interface{}) error {
	reqs, err := requests.Load(in)
	if err != nil {
		return fmt.Errorf("can't load requests from \"%s\"", in)
	}

	if n < 1 {
		n = len(reqs)
	}

	c := pep.NewClient(addr, nil)
	err = c.Connect()
	if err != nil {
		return fmt.Errorf("can't connect to %s: %s", addr, err)
	}
	defer c.Close()

	recs := make([]timing, n)

	p := v.(config).parallel
	if p > 0 {
		ch := make(chan int, p)

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

				recs[i].setSend()
				err := c.ModalValidate(req, res)
				if err != nil {
					recs[i].setError(err)
				} else {
					recs[i].setReceive()
				}
			}(i, reqs[i%len(reqs)])
		}

		wg.Wait()
	} else if p < 0 {
		var wg sync.WaitGroup
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func(i int, req pb.Request) {
				defer wg.Done()

				res := &pb.Response{}

				recs[i].setSend()
				err := c.ModalValidate(req, res)
				if err != nil {
					recs[i].setError(err)
				} else {
					recs[i].setReceive()
				}
			}(i, reqs[i%len(reqs)])
		}

		wg.Wait()
	} else {
		res := &pb.Response{}
		for i := 0; i < n; i++ {
			idx := i % len(reqs)
			req := reqs[idx]

			recs[i].setSend()
			err := c.ModalValidate(req, res)
			if err != nil {
				return fmt.Errorf("can't send request %d (%d): %s", idx, i, err)
			}
			recs[i].setReceive()
		}
	}

	tm := timings{
		Sends:    make([]int64, len(recs)),
		Receives: make([]int64, len(recs)),
		Pairs:    make([][]int64, len(recs)),
	}

	sort.Sort(bySend(recs))
	for i, t := range recs {
		tm.Sends[i] = t.s.UnixNano()
		if t.e != nil {
			tm.Pairs[i] = []int64{t.s.UnixNano()}
		} else {
			tm.Pairs[i] = []int64{
				t.s.UnixNano(),
				t.r.UnixNano(),
				t.r.UnixNano() - t.s.UnixNano(),
			}
		}
	}

	sort.Sort(byRecive(recs))
	for i, t := range recs {
		tm.Receives[i] = t.r.UnixNano()
	}

	b, err := json.MarshalIndent(tm, "", "  ")
	if err != nil {
		return fmt.Errorf("can't marshal timings to JSON: %s", err)
	}

	f := os.Stdout
	outName := "stdout"
	if len(out) > 0 {
		f, err = os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()

		outName = fmt.Sprintf("file %s", out)
	}

	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("can't dump JSON timings to %s: %s", outName, err)
	}

	return nil
}

type timing struct {
	s time.Time
	r time.Time
	e error
}

func (t *timing) setSend() {
	t.s = time.Now()
}

func (t *timing) setReceive() {
	t.r = time.Now()
}

func (t *timing) setError(err error) {
	t.e = err
}

type bySend []timing

func (s bySend) Len() int           { return len(s) }
func (s bySend) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s bySend) Less(i, j int) bool { return s[i].s.Before(s[j].s) }

type byRecive []timing

func (s byRecive) Len() int      { return len(s) }
func (s byRecive) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byRecive) Less(i, j int) bool {
	if s[i].e != nil && s[j].e != nil {
		return s[i].s.Before(s[j].s)
	}

	if s[i].e != nil {
		return false
	}

	if s[j].e != nil {
		return true
	}

	return s[i].r.Before(s[j].r)
}

type timings struct {
	Sends    []int64   `json:"sends"`
	Receives []int64   `json:"receives"`
	Pairs    [][]int64 `json:"pairs"`
}
