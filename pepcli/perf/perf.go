package perf

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
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

	tm := timings{
		Sends:    make([]int64, len(recs)),
		Receives: make([]int64, len(recs)),
		Pairs:    make([][3]int64, len(recs)),
	}

	for i, t := range recs {
		tm.Sends[i] = t.s.UnixNano()
		tm.Pairs[i][0] = t.s.UnixNano()
		tm.Pairs[i][1] = t.r.UnixNano()
		tm.Pairs[i][2] = tm.Pairs[i][1] - tm.Pairs[i][0]
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
}

func (t *timing) setSend() {
	t.s = time.Now()
}

func (t *timing) setReceive() {
	t.r = time.Now()
}

type byRecive []timing

func (s byRecive) Len() int           { return len(s) }
func (s byRecive) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s byRecive) Less(i, j int) bool { return s[i].r.Before(s[j].r) }

type timings struct {
	Sends    []int64    `json:"sends"`
	Receives []int64    `json:"receives"`
	Pairs    [][3]int64 `json:"pairs"`
}
