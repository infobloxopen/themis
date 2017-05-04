package policy

import (
	"fmt"
	"github.com/coredns/coredns/middleware"
	pdp "github.com/infobloxopen/themis/pdp-service"
	pep "github.com/infobloxopen/themis/pep"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
	"net"
	"reflect"
	"strings"
	"testing"
)

type TestMiddlewareHandler struct {
	status int
	err    error
}

func (f TestMiddlewareHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	var rr4 dns.RR
	rr4 = new(dns.A)
	rr4.(*dns.A).Hdr = dns.RR_Header{Name: "example.com", Rrtype: dns.TypeA, Class: dns.ClassINET}
	rr4.(*dns.A).A = net.ParseIP("10.10.10.10").To4()
	var rr6 dns.RR
	rr6 = new(dns.AAAA)
	rr6.(*dns.AAAA).Hdr = dns.RR_Header{Name: "example.com", Rrtype: dns.TypeAAAA, Class: dns.ClassINET}
	rr6.(*dns.AAAA).AAAA = net.ParseIP("21DA:D3:0:2F3B:2AA:FF:FE28:9C5A").To16()
	r.Answer = []dns.RR{rr4, rr6}

	w.WriteMsg(r)
	return f.status, f.err
}
func (f TestMiddlewareHandler) Name() string { return "handlerfunc" }

type TestCaseInfo struct {
	MiddlewareErr                      error
	MiddlewareStatus                   int
	ValidationResultMiddlewareResponse *pdp.Response
}

func TestHandlePermit(t *testing.T) {

	cases := []struct {
		c              TestCaseInfo
		expectedStatus int
		expectedErr    error
	}{
		{
			c: TestCaseInfo{
				MiddlewareErr:                      fmt.Errorf("Error"),
				MiddlewareStatus:                   dns.RcodeServerFailure,
				ValidationResultMiddlewareResponse: &pdp.Response{Effect: pdp.Response_DENY},
			},
			expectedStatus: dns.RcodeServerFailure,
			expectedErr:    fmt.Errorf("Error"),
		},
		{
			c: TestCaseInfo{
				MiddlewareErr:                      fmt.Errorf("Error"),
				MiddlewareStatus:                   dns.RcodeSuccess,
				ValidationResultMiddlewareResponse: &pdp.Response{Effect: pdp.Response_DENY},
			},
			expectedStatus: dns.RcodeSuccess,
			expectedErr:    fmt.Errorf("Error"),
		},
		{
			c: TestCaseInfo{
				MiddlewareErr:                      nil,
				MiddlewareStatus:                   dns.RcodeServerFailure,
				ValidationResultMiddlewareResponse: &pdp.Response{Effect: pdp.Response_DENY},
			},
			expectedStatus: dns.RcodeRefused,
			expectedErr:    nil,
		},
		{
			c: TestCaseInfo{
				MiddlewareErr:                      nil,
				MiddlewareStatus:                   dns.RcodeRefused,
				ValidationResultMiddlewareResponse: &pdp.Response{Effect: pdp.Response_PERMIT},
			},
			expectedStatus: dns.RcodeRefused,
			expectedErr:    nil,
		},
		{
			c: TestCaseInfo{
				MiddlewareErr:                      nil,
				MiddlewareStatus:                   dns.RcodeSuccess,
				ValidationResultMiddlewareResponse: &pdp.Response{Effect: pdp.Response_PERMIT},
			},
			expectedStatus: dns.RcodeSuccess,
			expectedErr:    nil,
		},
		{
			c: TestCaseInfo{
				MiddlewareErr:                      nil,
				MiddlewareStatus:                   dns.RcodeServerFailure,
				ValidationResultMiddlewareResponse: &pdp.Response{Effect: pdp.Response_PERMIT},
			},
			expectedStatus: dns.RcodeServerFailure,
			expectedErr:    nil,
		},
		{
			c: TestCaseInfo{
				MiddlewareErr:    nil,
				MiddlewareStatus: dns.RcodeSuccess,
				ValidationResultMiddlewareResponse: &pdp.Response{Effect: pdp.Response_PERMIT,
					Obligation: []*pdp.Attribute{{"redirect_to", "address", "221.228.88.194"}}},
			},
			expectedStatus: dns.RcodeSuccess,
			expectedErr:    nil,
		},
		{
			c: TestCaseInfo{
				MiddlewareErr:    nil,
				MiddlewareStatus: dns.RcodeServerFailure,
				ValidationResultMiddlewareResponse: &pdp.Response{Effect: pdp.Response_PERMIT,
					Obligation: []*pdp.Attribute{{"redirect_to", "address", "221.228.88.194"}}},
			},
			expectedStatus: dns.RcodeSuccess,
			expectedErr:    nil,
		},
		{
			c: TestCaseInfo{
				MiddlewareErr:    nil,
				MiddlewareStatus: dns.RcodeBadCookie,
				ValidationResultMiddlewareResponse: &pdp.Response{Effect: pdp.Response_PERMIT,
					Obligation: []*pdp.Attribute{{"redirect_to", "address", "221.228.88.194"}}},
			},
			expectedStatus: dns.RcodeSuccess,
			expectedErr:    nil,
		},
	}

	w := new(NewLocalResponseWriter)
	TestMiddleware := TestMiddlewareHandler{}
	p := new(PolicyMiddleware)
	c := pep.NewTestClient()
	p.pdp = c

	r := new(dns.Msg)
	r.SetReply(r)
	r.SetQuestion("www.example.com.", dns.TypeANY)
	ctx := context.Background()

	attrs := []*pdp.Attribute{{Id: "type", Type: "string", Value: "query"}}
	attrs = append(attrs, &pdp.Attribute{Id: "domain_name", Type: "domain", Value: strings.TrimRight("www.example.com", ".")})

	for _, ut := range cases {
		TestMiddleware.err = ut.c.MiddlewareErr
		TestMiddleware.status = ut.c.MiddlewareStatus
		p.Next = middleware.Handler(TestMiddleware)

		c.NextResponse = ut.c.ValidationResultMiddlewareResponse
		status, err := p.handlePermit(ctx, w, r, attrs)
		if !reflect.DeepEqual(err, ut.expectedErr) {
			t.Errorf("Expected err to be %q but it was %q", ut.expectedErr, err)
		}

		if ut.expectedStatus != status {
			t.Errorf("Expected status %q but got %q", ut.expectedStatus, status)
		}
	}

}
