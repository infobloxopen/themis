package policy

import (
	pdp "github.com/infobloxopen/themis/pdp-service"
)

type testClient struct {
	nextResponse   *pdp.Response
	nextResponseIP *pdp.Response
	errResponse    error
	errResponseIP  error
}

func newTestClientInit(nextResponse *pdp.Response, nextResponseIP *pdp.Response,
	errResponse error, errResponseIP error) *testClient {
	return &testClient{
		nextResponse:   nextResponse,
		nextResponseIP: nextResponseIP,
		errResponse:    errResponse,
		errResponseIP:  errResponseIP,
	}
}

func (c *testClient) Connect() error { return nil }
func (c *testClient) Close()         {}
func (c *testClient) Validate(request *pdp.Request) (*pdp.Response, error) {
	if request != nil {
		for _, a := range request.Attributes {
			if a.Id == "address" {
				if c.errResponseIP != nil {
					return nil, c.errResponseIP
				}
				if c.nextResponseIP != nil {
					return c.nextResponseIP, nil
				}
				continue
			}
		}
	}
	if c.errResponse != nil {
		return nil, c.errResponse
	}
	return c.nextResponse, nil
}
