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

func Exec(addr, in, out string, n, s int, v interface{}) error {
	reqs, err := requests.Load(in)
	if err != nil {
		return fmt.Errorf("can't load requests from \"%s\"", in)
	}

	if n < 1 {
		n = len(reqs)
	}

	opts := []pep.Option{}
	if s > 0 {
		opts = append(opts,
			pep.WithStreams(s),
		)
	}

	c := pep.NewClient(opts...)
	err = c.Connect(addr)
	if err != nil {
		return fmt.Errorf("can't connect to %s: %s", addr, err)
	}
	defer c.Close()

	validate := c.ModalValidate
	if s > 0 {
		validate = c.StreamValidate
	}

	recs, err := measurement(validate, n, v.(config).parallel, v.(config).limit, reqs)
	if err != nil {
		return err
	}

	return dump(recs, out)
}
