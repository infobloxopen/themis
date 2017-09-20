package test

import (
	"fmt"
	"os"

	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"
)

const (
	Name        = "test"
	Description = "tests given requests on PDP server"
)

func Exec(addr string, v interface{}) error {
	input := v.(config).input
	reqs, err := loadRequests(input)
	if err != nil {
		return fmt.Errorf("can't load requests from \"%s\"", input)
	}

	name := v.(config).output
	f := os.Stdout
	if len(name) > 0 {
		f, err = os.Create(name)
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

	for req := range reqs.parse(v.(config).count) {
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
