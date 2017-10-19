package policy

import (
	"testing"

	pdp "github.com/infobloxopen/themis/pdp-service"
)

func TestActionFromResponse(t *testing.T) {
	tests := []struct {
		resp     pdp.Response
		action   int
		redirect string
	}{
		{
			resp:     pdp.Response{Effect: pdp.Response_PERMIT},
			action:   typeAllow,
			redirect: "",
		},
		{
			resp:     pdp.Response{Effect: pdp.Response_INDETERMINATE},
			action:   typeInvalid,
			redirect: "",
		},
		{
			resp:     pdp.Response{Effect: pdp.Response_DENY},
			action:   typeBlock,
			redirect: "",
		},
		{
			resp: pdp.Response{Effect: pdp.Response_DENY, Obligation: []*pdp.Attribute{
				{Id: "redirect_to", Value: "10.10.10.10"},
			}},
			action:   typeRedirect,
			redirect: "10.10.10.10",
		},
		{
			resp: pdp.Response{Effect: pdp.Response_DENY, Obligation: []*pdp.Attribute{
				{Id: "refuse", Value: "bla-bla"},
			}},
			action:   typeBlock,
			redirect: "",
		},
		{
			resp: pdp.Response{Effect: pdp.Response_DENY, Obligation: []*pdp.Attribute{
				{Id: "refuse", Value: "true"},
			}},
			action:   typeRefuse,
			redirect: "",
		},
	}

	for _, test := range tests {
		a, r := actionFromResponse(&test.resp)
		if a != test.action {
			t.Errorf("Unexpected action: expected=%d, actual=%d", test.action, a)
		}
		strR := ""
		if r != nil {
			strR = r.GetValue()
		}
		if strR != test.redirect {
			t.Errorf("Unexpected redirect: expected=%q, actual=%q", test.redirect, strR)
		}
	}

	a, r := actionFromResponse(nil)
	if a != typeInvalid {
		t.Errorf("Unexpected action: expected=%d, actual=%d", typeInvalid, a)
	}
	if r != nil {
		t.Errorf("Unexpected redirect: expected=nil, actual=%q", r)
	}
}

func TestPolicyActionNegative(t *testing.T) {
	ah := newAttrHolder("test.com", "1")
	attrs := ah.attributes()
	for _, a := range attrs {
		if a.GetId() == "policy_action" {
			t.Errorf("Unexpected policy_action attribute")
		}
	}
}
