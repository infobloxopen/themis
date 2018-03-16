package policy

import (
	"testing"

	pdp "github.com/infobloxopen/themis/pdp-service"
)

func TestActionFromResponse(t *testing.T) {
	tests := []struct {
		resp     *pdp.Response
		action   byte
		redirect string
	}{
		{
			resp:     &pdp.Response{Effect: pdp.Response_PERMIT},
			action:   typeAllow,
			redirect: "",
		},
		{
			resp:     &pdp.Response{Effect: pdp.Response_INDETERMINATE},
			action:   typeInvalid,
			redirect: "",
		},
		{
			resp:     &pdp.Response{Effect: pdp.Response_DENY},
			action:   typeBlock,
			redirect: "",
		},
		{
			resp: &pdp.Response{Effect: pdp.Response_DENY, Obligation: []*pdp.Attribute{
				{Id: attrNameRedirectTo, Value: "10.10.10.10"},
			}},
			action:   typeRedirect,
			redirect: "10.10.10.10",
		},
		{
			resp: &pdp.Response{Effect: pdp.Response_DENY, Obligation: []*pdp.Attribute{
				{Id: attrNameRefuse, Value: "true"},
			}},
			action:   typeRefuse,
			redirect: "",
		},
		{
			resp: &pdp.Response{Effect: pdp.Response_PERMIT, Obligation: []*pdp.Attribute{
				{Id: attrNameLog, Value: ""},
			}},
			action:   typeLog,
			redirect: "",
		},
		{
			resp: &pdp.Response{Effect: pdp.Response_PERMIT, Obligation: []*pdp.Attribute{
				{Id: attrNameLog, Value: ""},
			}},
			action:   typeLog,
			redirect: "",
		},
		{
			resp: &pdp.Response{Effect: pdp.Response_DENY, Obligation: []*pdp.Attribute{
				{Id: attrNameDrop, Value: ""},
			}},
			action:   typeDrop,
			redirect: "",
		},
	}

	for i, test := range tests {
		ah := newAttrHolder("test.com", 1, "127.0.0.1", nil)
		ah.addResponse(test.resp, false)
		if ah.action != test.action {
			t.Errorf("Unexpected action in TC #%d: expected=%d, actual=%d", i, test.action, ah.action)
		}
		if ah.redirect != test.redirect {
			t.Errorf("Unexpected redirect in TC #%d: expected=%q, actual=%q", i, test.redirect, ah.redirect)
		}
	}
}

func TestAddResponse(t *testing.T) {
	tests := []struct {
		confAttrs     map[string]confAttrType
		ednsAttrs     []*pdp.Attribute
		resp          *pdp.Response
		expEdnsAttrs  []*pdp.Attribute
		expRespDomain []*pdp.Attribute
		expRespIp     []*pdp.Attribute
		expTransfer   []*pdp.Attribute
	}{
		{
			resp:          &pdp.Response{Effect: pdp.Response_PERMIT},
			expEdnsAttrs:  []*pdp.Attribute{},
			expRespDomain: []*pdp.Attribute{},
			expRespIp:     []*pdp.Attribute{},
			expTransfer:   []*pdp.Attribute{},
		},
		{
			confAttrs: map[string]confAttrType{
				"edns1": confAttrEdns,
			},
			ednsAttrs: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
			},
			resp: &pdp.Response{Effect: pdp.Response_PERMIT},
			expEdnsAttrs: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
			},
			expRespDomain: []*pdp.Attribute{},
			expRespIp:     []*pdp.Attribute{},
			expTransfer:   []*pdp.Attribute{},
		},
		{
			confAttrs: map[string]confAttrType{
				"edns1": confAttrEdns,
			},
			resp: &pdp.Response{Effect: pdp.Response_PERMIT, Obligation: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
			}},
			expEdnsAttrs: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
			},
			expRespDomain: []*pdp.Attribute{},
			expTransfer:   []*pdp.Attribute{},
		},
		{
			confAttrs: map[string]confAttrType{
				"edns1": confAttrEdns,
			},
			resp: &pdp.Response{Effect: pdp.Response_PERMIT, Obligation: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
			}},
			expEdnsAttrs: []*pdp.Attribute{},
			expRespIp: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
			},
			expTransfer: []*pdp.Attribute{},
		},
		{
			confAttrs: map[string]confAttrType{
				"edns1": confAttrEdns,
			},
			ednsAttrs: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
			},
			resp: &pdp.Response{Effect: pdp.Response_PERMIT, Obligation: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val2"},
			}},
			expEdnsAttrs: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
			},
			expRespDomain: []*pdp.Attribute{},
			expTransfer:   []*pdp.Attribute{},
		},
		{
			confAttrs: map[string]confAttrType{
				"edns1": confAttrEdns,
			},
			resp: &pdp.Response{Effect: pdp.Response_DENY, Obligation: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val2"},
			}},
			expEdnsAttrs: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val2"},
			},
			expRespDomain: []*pdp.Attribute{},
			expTransfer:   []*pdp.Attribute{},
		},
		{
			confAttrs: map[string]confAttrType{
				"trans1": confAttrTransfer,
			},
			resp: &pdp.Response{Effect: pdp.Response_PERMIT, Obligation: []*pdp.Attribute{
				{Id: "trans1", Value: "trans1Val"},
			}},
			expEdnsAttrs: []*pdp.Attribute{},
			expRespIp: []*pdp.Attribute{
				{Id: "trans1", Value: "trans1Val"},
			},
			expTransfer: []*pdp.Attribute{},
		},
		{
			confAttrs: map[string]confAttrType{
				"trans1": confAttrTransfer,
			},
			resp: &pdp.Response{Effect: pdp.Response_PERMIT, Obligation: []*pdp.Attribute{
				{Id: "trans1", Value: "trans1Val"},
			}},
			expEdnsAttrs: []*pdp.Attribute{},
			expRespDomain: []*pdp.Attribute{
				{Id: "trans1", Value: "trans1Val"},
			},
			expTransfer: []*pdp.Attribute{
				{Id: "trans1", Value: "trans1Val"},
			},
		},
		{
			confAttrs: map[string]confAttrType{
				"trans1": confAttrTransfer,
			},
			resp: &pdp.Response{Effect: pdp.Response_DENY, Obligation: []*pdp.Attribute{
				{Id: "trans1", Value: "trans1Val"},
			}},
			expEdnsAttrs: []*pdp.Attribute{},
			expRespDomain: []*pdp.Attribute{
				{Id: "trans1", Value: "trans1Val"},
			},
			expTransfer: []*pdp.Attribute{},
		},
		{
			confAttrs: map[string]confAttrType{
				"edns1":  confAttrEdns,
				"trans1": confAttrTransfer,
			},
			ednsAttrs: []*pdp.Attribute{},
			resp: &pdp.Response{Effect: pdp.Response_PERMIT, Obligation: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
				{Id: "trans1", Value: "trans1Val"},
				{Id: "other1", Value: "other1Val"},
			}},
			expEdnsAttrs: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
			},
			expRespDomain: []*pdp.Attribute{
				{Id: "trans1", Value: "trans1Val"},
				{Id: "other1", Value: "other1Val"},
			},
			expTransfer: []*pdp.Attribute{
				{Id: "trans1", Value: "trans1Val"},
			},
		},
		{
			confAttrs: map[string]confAttrType{
				"edns1":  confAttrEdns,
				"trans1": confAttrTransfer,
			},
			ednsAttrs: []*pdp.Attribute{},
			resp: &pdp.Response{Effect: pdp.Response_PERMIT, Obligation: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
				{Id: "trans1", Value: "trans1Val"},
				{Id: "other1", Value: "other1Val"},
			}},
			expEdnsAttrs: []*pdp.Attribute{},
			expRespIp: []*pdp.Attribute{
				{Id: "edns1", Value: "ends1Val"},
				{Id: "trans1", Value: "trans1Val"},
				{Id: "other1", Value: "other1Val"},
			},
			expTransfer: []*pdp.Attribute{},
		},
	}

	for i, test := range tests {
		if test.expRespDomain != nil {
			ah := newAttrHolder("test.com", 1, "127.0.0.1", test.confAttrs)
			if test.ednsAttrs != nil {
				ah.attrsReqDomain = append(ah.attrsReqDomain, test.ednsAttrs...)
			}
			ah.addResponse(test.resp, false)
			checkAttrs(t, "check respDomain attrs", i, ah.attrsRespDomain, test.expRespDomain)
			if test.expEdnsAttrs != nil {
				checkAttrs(t, "check edns attrs", i, ah.attrsReqDomain[ah.attrsEdnsStart:], test.expEdnsAttrs)
			}
			if test.expTransfer != nil {
				checkAttrs(t, "check transfer attrs", i, ah.attrsTransfer, test.expTransfer)
			}
		}
		if test.expRespIp != nil {
			ah := newAttrHolder("test.com", 1, "127.0.0.1", test.confAttrs)
			if test.ednsAttrs != nil {
				ah.attrsReqDomain = append(ah.attrsReqDomain, test.ednsAttrs...)
			}
			ah.addResponse(test.resp, true)
			checkAttrs(t, "check respRespip attrs", i, ah.attrsRespRespip, test.expRespIp)
			if test.expEdnsAttrs != nil {
				checkAttrs(t, "check edns attrs", i, ah.attrsReqDomain[ah.attrsEdnsStart:], test.expEdnsAttrs)
			}
			if test.expTransfer != nil {
				checkAttrs(t, "check transfer attrs", i, ah.attrsTransfer, test.expTransfer)
			}
		}
	}
}

func TestNilResponse(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("addResponse() did not panic for nil response")
		}
	}()

	ah := newAttrHolder("test.com", 1, "127.0.0.1", nil)
	ah.addResponse(nil, false)
}

func TestAllowActionAfterLogAction(t *testing.T) {
	ah := newAttrHolder("test.com", 1, "127.0.0.1", nil)
	ah.addResponse(&pdp.Response{Effect: pdp.Response_PERMIT,
		Obligation: []*pdp.Attribute{{Id: "log"}}}, false)
	ah.addResponse(&pdp.Response{Effect: pdp.Response_PERMIT}, true)
	if ah.action != typeLog {
		t.Errorf("Unexpected action: expected=%d, actual=%d", typeLog, ah.action)
	}
}

func checkAttrs(t *testing.T, msg string, testNo int, actual []*pdp.Attribute, expected []*pdp.Attribute) {
	if len(actual) != len(expected) {
		t.Errorf("Test #%d: %s - expected %d attributes, found %d", testNo, msg, len(expected), len(actual))
	}

checkAttr:
	for _, a := range actual {
		for _, e := range expected {
			if e.Id == a.Id {
				if a.Value != e.Value {
					t.Errorf("Test #%d: %s - attribute %q - expected=%q, actual=%q", testNo, msg, e.Id, e.Value, a.Value)
					return
				}
				continue checkAttr
			}
		}
		t.Errorf("Test #%d: %s - unexpected attribute found %q=%q", testNo, msg, a.Id, a.Value)
	}
}
