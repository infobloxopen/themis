package policy

import (
	"errors"
	"fmt"
	"testing"

	"github.com/coredns/coredns/middleware"
	"github.com/coredns/coredns/middleware/pkg/dnsrecorder"
	"github.com/coredns/coredns/middleware/test"

	pdp "github.com/infobloxopen/themis/pdp-service"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

var (
	errFakePdp      = errors.New("fake PDP error")
	errFakeResolver = errors.New("fake Resolver error")
)

func TestPolicy(t *testing.T) {
	pm := PolicyMiddleware{Next: handler()}

	tests := []struct {
		query      string
		queryType  uint16
		response   *pdp.Response
		responseIP *pdp.Response
		errResp    error
		errRespIP  error
		status     int
		err        error
	}{
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			errResp:   errFakePdp,
			status:    dns.RcodeServerFailure,
			err:       errFakePdp,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.Response_PERMIT},
			errRespIP: errFakePdp,
			status:    dns.RcodeServerFailure,
			err:       errFakePdp,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.Response_DENY},
			status:    dns.RcodeNameError,
			err:       nil,
		},
		{
			query:      "test.com.",
			queryType:  dns.TypeA,
			response:   &pdp.Response{Effect: pdp.Response_PERMIT},
			responseIP: &pdp.Response{Effect: pdp.Response_PERMIT},
			status:     dns.RcodeSuccess,
			err:        nil,
		},
		{
			query:      "test.com.",
			queryType:  dns.TypeA,
			response:   &pdp.Response{Effect: pdp.Response_PERMIT},
			responseIP: &pdp.Response{Effect: pdp.Response_DENY},
			status:     dns.RcodeNameError,
			err:        nil,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response: &pdp.Response{Effect: pdp.Response_DENY,
				Obligation: []*pdp.Attribute{{"redirect_to", "string", "221.228.88.194"}}},
			status: dns.RcodeSuccess,
			err:    nil,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response: &pdp.Response{Effect: pdp.Response_DENY,
				Obligation: []*pdp.Attribute{{"redirect_to", "string", "redirect.biz"}}},
			status: dns.RcodeSuccess,
			err:    nil,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.Response_PERMIT},
			responseIP: &pdp.Response{Effect: pdp.Response_DENY,
				Obligation: []*pdp.Attribute{{"redirect_to", "string", "221.228.88.194"}}},
			status: dns.RcodeSuccess,
			err:    nil,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.Response_PERMIT},
			responseIP: &pdp.Response{Effect: pdp.Response_DENY,
				Obligation: []*pdp.Attribute{{"redirect_to", "string", "redirect.biz"}}},
			status: dns.RcodeSuccess,
			err:    nil,
		},
		{
			query:     "test.com.",
			queryType: dns.TypeA,
			response: &pdp.Response{Effect: pdp.Response_DENY,
				Obligation: []*pdp.Attribute{{"redirect_to", "string", "test.net"}}},
			status: dns.RcodeServerFailure,
			err:    errFakeResolver,
		},
		{
			query:     "test.net.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.Response_PERMIT},
			status:    dns.RcodeServerFailure,
			err:       errFakeResolver,
		},
		{
			query:     "test.org.",
			queryType: dns.TypeAAAA,
			response:  &pdp.Response{Effect: pdp.Response_DENY},
			status:    dns.RcodeNameError,
			err:       nil,
		},
		{
			query:     "test.org.",
			queryType: dns.TypeAAAA,
			response:  &pdp.Response{Effect: pdp.Response_PERMIT},
			status:    dns.RcodeSuccess,
			err:       nil,
		},
		{
			query:     "test.org.",
			queryType: dns.TypeAAAA,
			response: &pdp.Response{Effect: pdp.Response_DENY,
				Obligation: []*pdp.Attribute{{"redirect_to", "string", "redirect.net"}}},
			status: dns.RcodeSuccess,
			err:    nil,
		},
		{
			query:     "test.org.",
			queryType: dns.TypeA,
			response: &pdp.Response{Effect: pdp.Response_DENY,
				Obligation: []*pdp.Attribute{{"refuse", "string", "true"}}},
			status: dns.RcodeRefused,
			err:    nil,
		},
		{
			query:     "nxdomain.org.",
			queryType: dns.TypeA,
			response:  &pdp.Response{Effect: pdp.Response_PERMIT},
			status:    dns.RcodeNameError,
			err:       nil,
		},
	}

	rec := dnsrecorder.New(&test.ResponseWriter{})

	for _, test := range tests {
		req := new(dns.Msg)
		req.SetQuestion(test.query, test.queryType)
		// add ENDS options to request
		o := new(dns.OPT)
		o.Hdr.Name = "."
		o.Hdr.Rrtype = dns.TypeOPT
		e := new(dns.EDNS0_LOCAL)
		e.Code = 0xfffa
		e.Data = []byte("4e7e318384088e7d4f3dbc96219ee5d4")
		o.Option = append(o.Option, e)
		req.Extra = append(req.Extra, o)
		// Init test mock client
		pm.pdp = newTestClientInit(test.response, test.responseIP, test.errResp, test.errRespIP)
		// Add EDNS mapping
		pm.AddEDNS0Map("0xfffa", "client_id", "hex", "string", "0", "32")
		pm.AddEDNS0Map("0xfffb", "client_src", "address", "string", "", "")
		pm.AddEDNS0Map("0xfffc", "client_mac", "bytes", "string", "", "")
		// Handle request
		status, err := pm.ServeDNS(context.TODO(), rec, req)
		// Check status
		if test.status != status {
			t.Errorf("Expected status %q but got %q\n", test.status, status)
		}
		// Check error
		if test.err != err {
			t.Errorf("Expected error %v but got %v\n", test.err, err)
		}
	}

}

func handler() middleware.Handler {
	return middleware.HandlerFunc(func(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
		q := r.Question[0].Name
		switch q {
		case "test.com.":
			r.Answer = []dns.RR{
				test.A("test.com.		600	IN	A			10.240.0.1"),
			}
		case "test.org.":
			r.Answer = []dns.RR{
				test.AAAA("test.org.	600	IN	AAAA		21DA:D3:0:2F3B:2AA:FF:FE28:9C5A"),
			}
		case "redirect.biz.":
			r.Answer = []dns.RR{
				test.A("redirect.biz.	600	IN	A			221.228.88.194"),
			}
		case "redirect.net.":
			r.Answer = []dns.RR{
				test.AAAA("redirect.net.	600	IN	AAAA		2001:db8:0:200:0:0:0:7"),
			}
		case "nxdomain.org.":
			w.WriteMsg(r)
			return dns.RcodeNameError, nil
		default:
			return dns.RcodeServerFailure, errFakeResolver
		}
		w.WriteMsg(r)
		return dns.RcodeSuccess, nil
	})
}

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
func (c *testClient) Validate(ctx context.Context, in, out interface{}) error {
	if in != nil {
		p := in.(pdp.Request)
		for _, a := range p.Attributes {
			if a.Id == "address" {
				if c.errResponseIP != nil {
					return c.errResponseIP
				}
				if c.nextResponseIP != nil {
					return fillResponse(c.nextResponseIP, out)
				}
				continue
			}
		}
	}
	if c.errResponse != nil {
		return c.errResponse
	}
	return fillResponse(c.nextResponse, out)
}

func (c *testClient) ModalValidate(in, out interface{}) error {
	return c.Validate(context.Background(), in, out)
}

func fillResponse(in *pdp.Response, out interface{}) error {
	r, ok := out.(*Response)
	if !ok {
		return fmt.Errorf("testClient can only translate response to *Response type but got %T", out)
	}

	r.Permit = in.Effect == pdp.Response_PERMIT

	for _, attr := range in.Obligation {
		switch attr.Id {
		case "redirect_to":
			r.Redirect = attr.Value

		case "policy_id":
			r.PolicyID = attr.Value

		case "refuse":
			r.Refuse = attr.Value
		}
	}

	return nil
}
