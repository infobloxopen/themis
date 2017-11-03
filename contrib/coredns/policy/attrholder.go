package policy

import (
	"log"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	pdp "github.com/infobloxopen/themis/pdp-service"
)

var actionConvDnstap [actCount]string

func init() {
	actionConvDnstap[typeInvalid] = "0"  // dnstap.PolicyAction_INVALID
	actionConvDnstap[typeRefuse] = "5"   // dnstap.PolicyAction_REFUSE
	actionConvDnstap[typeAllow] = "2"    // dnstap.PolicyAction_PASSTHROUGH
	actionConvDnstap[typeRedirect] = "4" // dnstap.PolicyAction_REDIRECT
	actionConvDnstap[typeBlock] = "3"    // dnstap.PolicyAction_NXDOMAIN
	actionConvDnstap[typeLog] = "2"      // dnstap.PolicyAction_PASSTHROUGH
}

type attrHolder struct {
	attrsReqDomain  []*pdp.Attribute
	attrsRespDomain []*pdp.Attribute
	attrsReqRespip  []*pdp.Attribute
	attrsRespRespip []*pdp.Attribute
	action          byte
	redirect        string
}

func newAttrHolder(qName string, qType uint16) *attrHolder {
	ret := &attrHolder{
		attrsReqDomain: make([]*pdp.Attribute, 3, 8),
		action:         typeInvalid,
	}
	ret.attrsReqDomain[0] = &pdp.Attribute{Id: "type", Type: "string", Value: "query"}
	ret.attrsReqDomain[1] = &pdp.Attribute{Id: "domain_name", Type: "domain", Value: strings.TrimRight(qName, ".")}
	ret.attrsReqDomain[2] = &pdp.Attribute{Id: "dns_qtype", Type: "string", Value: strconv.FormatUint(uint64(qType), 16)}
	return ret
}

func (ah *attrHolder) makeReqRespip(addr string) {
	policyID := ""
	for _, item := range ah.attrsRespDomain {
		if item.Id == "policy_id" {
			policyID = item.Value
			break
		}
	}

	ah.attrsReqRespip = []*pdp.Attribute{
		{Id: "type", Type: "string", Value: "response"},
		{Id: "policy_id", Type: "string", Value: policyID},
		{Id: "address", Type: "address", Value: addr},
	}
}

func (ah *attrHolder) addResponse(r *pdp.Response, respip bool) {
	if !respip {
		ah.attrsRespDomain = r.Obligations
	} else {
		ah.attrsRespRespip = r.Obligations
	}

	switch r.Effect {
	case pdp.PERMIT:
		for _, item := range r.Obligations {
			if item.Id == "log" {
				ah.action = typeLog
				return
			}
		}
		// don't overwrite "Log" action from previous validation
		if ah.action != typeLog {
			ah.action = typeAllow
		}
		return
	case pdp.DENY:
		for _, item := range r.Obligations {
			switch item.Id {
			case "refuse":
				ah.action = typeRefuse
				return
			case "redirect_to":
				ah.action = typeRedirect
				ah.redirect = item.Value
				return
			}
		}
		ah.action = typeBlock
	default:
		log.Printf("[ERROR] PDP Effect: %s", pdp.EffectName(r.Effect))
		ah.action = typeInvalid
	}
	return
}

func (ah *attrHolder) convertAttrs() []*dnstap.DnstapAttribute {
	lenAttrsReqDomain := len(ah.attrsReqDomain) - 1
	lenAttrsRespDomain := len(ah.attrsRespDomain)
	lenAttrsReqRespip := len(ah.attrsReqRespip)
	if lenAttrsReqRespip > 0 {
		lenAttrsReqRespip -= 2
	}
	lenAttrsRespRespip := len(ah.attrsRespRespip)
	length := lenAttrsReqDomain + lenAttrsRespDomain +
		lenAttrsReqRespip + lenAttrsRespRespip + 1
	out := make([]*dnstap.DnstapAttribute, length)
	i := 0
	id := "policy_action"
	out[i] = &dnstap.DnstapAttribute{Id: id, Value: actionConvDnstap[ah.action]}
	i++
	for j := 0; j < lenAttrsReqDomain; j++ {
		out[i] = &dnstap.DnstapAttribute{
			Id:    ah.attrsReqDomain[j+1].Id,
			Value: ah.attrsReqDomain[j+1].Value,
		}
		i++
	}
	for j := 0; j < lenAttrsRespDomain; j++ {
		out[i] = &dnstap.DnstapAttribute{
			Id:    ah.attrsRespDomain[j].Id,
			Value: ah.attrsRespDomain[j].Value,
		}
		i++
	}
	for j := 0; j < lenAttrsReqRespip; j++ {
		out[i] = &dnstap.DnstapAttribute{
			Id:    ah.attrsReqRespip[j+2].Id,
			Value: ah.attrsReqRespip[j+2].Value,
		}
		i++
	}
	for j := 0; j < lenAttrsRespRespip; j++ {
		out[i] = &dnstap.DnstapAttribute{
			Id:    ah.attrsRespRespip[j].Id,
			Value: ah.attrsRespRespip[j].Value,
		}
		i++
	}
	return out
}
