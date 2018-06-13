package pep

import (
	"golang.org/x/net/context"

	ps "github.com/infobloxopen/themis/pdpserver/server"
)

type builtinClient struct {
	s *ps.Server
}

func NewBuiltinClient(policyFile string, contentFiles []string) *builtinClient {
	s := ps.NewIntegratedServer(policyFile, contentFiles)
	return &builtinClient{
		s: s,
	}
}

func (c *builtinClient) Connect(addr string) error {
	return nil
}

func (c *builtinClient) Close() {}

func (c *builtinClient) Validate(in, out interface{}) error {
	if c.s == nil {
		return ErrorNotConnected
	}

	req, err := makeRequest(in)
	if err != nil {
		return err
	}

	res, err := c.s.Validate(context.Background(), &req)
	if err != nil {
		return err
	}

	return fillResponse(res, out)
}
