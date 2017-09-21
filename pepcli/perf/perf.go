package perf

import (
	"fmt"

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

	recs, err := measurement(c, n, v.(config).parallel, v.(config).limit, reqs)
	if err != nil {
		return err
	}

	return dump(recs, out)
}
