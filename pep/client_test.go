package pep

import (
	pdp "github.com/infobloxopen/themis/pdp-service"
	context "golang.org/x/net/context"
	"testing"
)

func TestTestClient(t *testing.T) {
	c := NewTestClient()

	c.NextResponse = &pdp.Response{Effect: pdp.Response_PERMIT,
		Obligation: []*pdp.Attribute{{"String", "string", "test"}}}

	res := &TestResponseStruct{}

	c.Validate(context.TODO(), nil, res)

	if !res.Effect {
		t.Errorf("Expected PERMIT but got %v", res.Effect)
	}

	if res.String != "test" {
		t.Errorf("Expected 'test' but got %q for res.String", res.String)
	}
}
