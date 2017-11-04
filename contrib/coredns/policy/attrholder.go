package policy

import (
	"log"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/contrib/coredns/policy/pb"
	pdp "github.com/infobloxopen/themis/pdp-service"
)

var actionConvDnstap [actCount]string

func init() {
	actionConvDnstap[typeInvalid] = "0"  // pb.PolicyAction_INVALID
	actionConvDnstap[typeRefuse] = "5"   // pb.PolicyAction_REFUSE
	actionConvDnstap[typeAllow] = "2"    // pb.PolicyAction_PASSTHROUGH
	actionConvDnstap[typeRedirect] = "4" // pb.PolicyAction_REDIRECT
	actionConvDnstap[typeBlock] = "3"    // pb.PolicyAction_NXDOMAIN
	actionConvDnstap[typeLog] = "2"      // pb.PolicyAction_PASSTHROUGH
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

func (ah *attrHolder) convertAttrs() []*pb.DnstapAttribute {
	lenAttrsReqDomain := len(ah.attrsReqDomain)
	lenAttrsRespDomain := len(ah.attrsRespDomain)
	lenAttrsReqRespip := len(ah.attrsReqRespip)
	lenAttrsRespRespip := len(ah.attrsRespRespip)
	length := lenAttrsReqDomain + lenAttrsRespDomain + lenAttrsReqRespip + lenAttrsRespRespip
	if lenAttrsReqRespip > 0 {
		length -= 2
	}
	out := make([]*pb.DnstapAttribute, length)
	i := 0
	for j := 1; j < lenAttrsReqDomain; j++ {
		out[i] = &pb.DnstapAttribute{
			Id:    ah.attrsReqDomain[j].Id,
			Value: ah.attrsReqDomain[j].Value,
		}
		i++
	}
	for j := 0; j < lenAttrsRespDomain; j++ {
		out[i] = &pb.DnstapAttribute{
			Id:    ah.attrsRespDomain[j].Id,
			Value: ah.attrsRespDomain[j].Value,
		}
		i++
	}
	for j := 2; j < lenAttrsReqRespip; j++ {
		out[i] = &pb.DnstapAttribute{
			Id:    ah.attrsReqRespip[j].Id,
			Value: ah.attrsReqRespip[j].Value,
		}
		i++
	}
	for j := 0; j < lenAttrsRespRespip; j++ {
		out[i] = &pb.DnstapAttribute{
			Id:    ah.attrsRespRespip[j].Id,
			Value: ah.attrsRespRespip[j].Value,
		}
		i++
	}
	out[i] = &pb.DnstapAttribute{Id: "policy_action", Value: actionConvDnstap[ah.action]}
	return out
}
