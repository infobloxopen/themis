package test

import (
	"fmt"
	"os"

	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"
)

const (
	Name        = "test"
	Description = "evaluates given requests on PDP server"
)

func Exec(addr, in, out string, n int, v interface{}) error {
	reqs, err := loadRequests(in)
	if err != nil {
		return fmt.Errorf("can't load requests from \"%s\"", in)
	}

	f := os.Stdout
	if len(out) > 0 {
		f, err = os.Create(out)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	c := pep.NewClient(addr, nil)
	err = c.Connect()
	if err != nil {
		return fmt.Errorf("can't connect to %s: %s", addr, err)
	}
	defer c.Close()

	for req := range reqs.parse(n) {
		if req.err != nil {
			return fmt.Errorf("don't understand request %d: %s", req.position, req.err)
		}

		res := &pb.Response{}
		err := c.ModalValidate(req.request, res)
		if err != nil {
			return fmt.Errorf("can't send request %d (%d): %s", req.index, req.position, err)
		}

		err = dumpResponse(res, f)
		if err != nil {
			return fmt.Errorf("can't dump response for reqiest %d (%d): %s", req.index, req.position, err)
		}
	}

	return nil
}
