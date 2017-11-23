package policy

import (
	"fmt"

	pdp "github.com/infobloxopen/themis/pdp-service"
)

type testClient struct {
	nextResponse *pdp.Response
	errResponse  error
}

func newTestClientInit(nextResponse *pdp.Response, errResponse error) *testClient {
	return &testClient{
		nextResponse: nextResponse,
		errResponse:  errResponse,
	}
}

func (c *testClient) Connect(addr string) error { return nil }
func (c *testClient) Close()                    {}
func (c *testClient) Validate(in, out interface{}) error {
	if c.errResponse != nil {
		return c.errResponse
	}
	return fillResponse(c.nextResponse, out)
}

func fillResponse(in *pdp.Response, out interface{}) error {
	r, ok := out.(*pdp.Response)
	if !ok {
		return fmt.Errorf("testClient can only translate response to *Response type but got %T", out)
	}
	r.Effect = in.Effect
	r.Obligation = in.Obligation
	return nil
}
