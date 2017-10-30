package policy

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	pdp "github.com/infobloxopen/themis/pdp-service"
)

var actionConv [actCount]string

func init() {
	actionConv[typeInvalid] = "0"
	actionConv[typeRefuse] = fmt.Sprintf("%x", int32(dnstap.PolicyAction_REFUSE))
	actionConv[typeAllow] = fmt.Sprintf("%x", int32(dnstap.PolicyAction_PASSTHROUGH))
	actionConv[typeRedirect] = fmt.Sprintf("%x", int32(dnstap.PolicyAction_REDIRECT))
	actionConv[typeBlock] = fmt.Sprintf("%x", int32(dnstap.PolicyAction_NXDOMAIN))
}

// correct sequence of func calls
// 1 newAttrHolder
// 2 (optionally) addAttr
// 3 request
// 4 addResponse
// 5 (optionally) addAttr
// 6 (optionally) request
// 7 (optionally) addResponse
//
// resp1, resp2, attributes (in any order)
type attrHolder struct {
	attrs    []*pdp.Attribute
	redirect *pdp.Attribute
	effect1  pdp.Response_Effect
	effect2  pdp.Response_Effect
	action   int
	typeInd  int
	resp1Beg int
	resp1End int
	resp2Beg int
	resp2End int
}

func newAttrHolder(qName string, qType uint16) *attrHolder {
	attrs := make([]*pdp.Attribute, 2, 32)
	attrs[0] = &pdp.Attribute{Id: "dns_qtype", Type: "string", Value: strconv.FormatUint(uint64(qType), 16)}
	attrs[1] = &pdp.Attribute{Id: "domain_name", Type: "domain", Value: strings.TrimRight(qName, ".")}
	return &attrHolder{attrs: attrs, action: typeInvalid}
}

func (ah *attrHolder) addAttr(a *pdp.Attribute) {
	ah.attrs = append(ah.attrs, a)
}

func (ah *attrHolder) addAttrs(a []*pdp.Attribute) {
	ah.attrs = append(ah.attrs, a...)
}

func (ah *attrHolder) setTypeAttr() {
	if len(ah.attrs) == 0 {
		panic("adding type attribute to empty list")
	}
	if ah.typeInd == 0 {
		ah.typeInd = len(ah.attrs)
		t := pdp.Attribute{Id: "type", Type: "string", Value: "query"}
		ah.attrs = append(ah.attrs, &t)
	} else {
		ah.attrs[ah.typeInd].Value = "response"
	}
}

func (ah *attrHolder) addAddress(val string) {
	aa := pdp.Attribute{Id: "address", Type: "address", Value: val}
	ah.addAttr(&aa)
}

func (ah *attrHolder) request() []*pdp.Attribute {
	ah.setTypeAttr()
	beg := 1 // skip "dns_qtype" since PDP doesn't need it
	if ah.resp1Beg != 0 {
		beg = ah.typeInd
	}
	return ah.attrs[beg:]
}

func (ah *attrHolder) addResponse(r *pdp.Response) {
	a := r.Obligation
	if ah.resp1Beg == 0 {
		ah.resp1Beg = len(ah.attrs)
		ah.resp1End = ah.resp1Beg + len(a)
		ah.effect1 = r.Effect
	} else {
		ah.resp2Beg = len(ah.attrs)
		ah.resp2End = ah.resp2Beg + len(a)
		ah.effect2 = r.Effect
	}
	ah.attrs = append(ah.attrs, a...)
	ah.action, ah.redirect = actionFromResponse(r)
}

func (ah *attrHolder) resp1() []*pdp.Attribute {
	return ah.attrs[ah.resp1Beg:ah.resp1End]
}

func (ah *attrHolder) resp2() []*pdp.Attribute {
	return ah.attrs[ah.resp2Beg:ah.resp2End]
}

func (ah *attrHolder) attributes() []*pdp.Attribute {
	if ah.action == typeInvalid {
		return ah.attrs
	}
	actAttr := pdp.Attribute{Id: "policy_action", Type: "string", Value: actionConv[ah.action]}
	return append(ah.attrs, &actAttr)
}

// actionFromResponse returns action and optionally pointer to redirect attribute
func actionFromResponse(resp *pdp.Response) (int, *pdp.Attribute) {
	if resp == nil {
		log.Printf("[ERROR] PDP response pointer is nil")
		return typeInvalid, nil
	}
	if resp.Effect == pdp.Response_PERMIT {
		return typeAllow, nil
	}
	if resp.Effect == pdp.Response_DENY {
		for _, item := range resp.Obligation {
			switch item.Id {
			case "refuse":
				return typeRefuse, nil
			case "redirect_to":
				return typeRedirect, item
			}
		}
		return typeBlock, nil
	}
	log.Printf("[ERROR] PDP Effect: %s", resp.Effect)
	return typeInvalid, nil
}
