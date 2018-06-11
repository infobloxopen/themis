package pep

import (
	"golang.org/x/net/context"

	ps "github.com/infobloxopen/themis/pdpserver/server"
)

type integratedClient struct {
	s *ps.Server
}

func NewIntegratedClient(policyFile string, contentFiles []string) *integratedClient {
	s := ps.NewIntegratedServer(policyFile, contentFiles)
	return &integratedClient{
		s: s,
	}
}

func (c *integratedClient) Connect(addr string) error {
	return nil
}

func (c *integratedClient) Close() {}

func (c *integratedClient) Validate(in, out interface{}) error {
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
