package policy

import (
	"context"
	pb "github.com/infobloxopen/policy-box/pdp-service"
	"github.com/infobloxopen/policy-box/pep"
	"github.com/miekg/dns"
	"testing"
)

// This function unit tests handlePermit function
// Inputs are t *testing.T
// Output is update in t
func TestHandlePermit(t *testing.T) {
	p := new(PolicyMiddleware)
	p.pdp = new(pep.Client)
	tw := new(NewLocalResponseWriter)
	ctx := context.Background()

	r := new(dns.Msg)
	r.SetQuestion("example.com.", dns.TypeANY)
	attrs := []*pb.Attribute{&pb.Attribute{Id: "type", Type: "string", Value: "query"}}

	status, err := p.handlePermit(ctx, tw, r, attrs)

	if err == nil {
		t.Fatalf("Expected status dns.RcodeRefused and error != nil but got %#v,%#v", status, err)
	}

}
