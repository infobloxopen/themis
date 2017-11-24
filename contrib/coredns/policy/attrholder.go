package policy

import (
	"log"
	"strconv"
	"strings"

	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
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
	transfer        map[string]struct{}
	attrsReqDomain  []*pdp.Attribute
	attrsRespDomain []*pdp.Attribute
	attrsReqRespip  []*pdp.Attribute
	attrsRespRespip []*pdp.Attribute
	action          byte
	redirect        string
}

func newAttrHolder(qName string, qType uint16, transfer map[string]struct{}) *attrHolder {
	ret := &attrHolder{
		transfer:       transfer,
		attrsReqDomain: make([]*pdp.Attribute, 3, 8),
		action:         typeInvalid,
	}
	ret.attrsReqDomain[0] = &pdp.Attribute{Id: AttrNameType, Type: "string", Value: TypeValueQuery}
	ret.attrsReqDomain[1] = &pdp.Attribute{Id: AttrNameDomainName, Type: "domain", Value: strings.TrimRight(qName, ".")}
	ret.attrsReqDomain[2] = &pdp.Attribute{Id: AttrNameDNSQtype, Type: "string", Value: strconv.FormatUint(uint64(qType), 16)}
	return ret
}

func (ah *attrHolder) makeReqRespip(addr string) {
	ah.attrsReqRespip = []*pdp.Attribute{
		{Id: AttrNameType, Type: "string", Value: TypeValueResponse},
		{Id: AttrNameAddress, Type: "address", Value: addr},
	}

	for _, item := range ah.attrsRespDomain {
		if _, ok := ah.transfer[item.Id]; ok {
			ah.attrsReqRespip = append(ah.attrsReqRespip, item)
		}
	}
}

func (ah *attrHolder) addResponse(r *pdp.Response, respip bool) {
	if !respip {
		ah.attrsRespDomain = r.Obligation
	} else {
		ah.attrsRespRespip = r.Obligation
	}

	switch r.Effect {
	case pdp.Response_PERMIT:
		for _, item := range r.Obligation {
			if item.Id == AttrNameLog {
				ah.action = typeLog
				return
			}
		}
		// don't overwrite "Log" action from previous validation
		if ah.action != typeLog {
			ah.action = typeAllow
		}
		return
	case pdp.Response_DENY:
		for _, item := range r.Obligation {
			switch item.Id {
			case AttrNameRefuse:
				ah.action = typeRefuse
				return
			case AttrNameRedirectTo:
				ah.action = typeRedirect
				ah.redirect = item.Value
				return
			}
		}
		ah.action = typeBlock
	default:
		log.Printf("[ERROR] PDP Effect: %s", r.Effect)
		ah.action = typeInvalid
	}
	return
}

func (ah *attrHolder) convertAttrs() []*pb.DnstapAttribute {
	lenAttrsReqDomain := len(ah.attrsReqDomain)
	lenAttrsRespDomain := len(ah.attrsRespDomain)
	lenAttrsReqRespip := len(ah.attrsReqRespip)
	if lenAttrsReqRespip > 0 {
		lenAttrsReqRespip = 1
	}
	lenAttrsRespRespip := len(ah.attrsRespRespip)
	length := lenAttrsReqDomain + lenAttrsRespDomain + lenAttrsReqRespip + lenAttrsRespRespip + 1
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
	if lenAttrsReqRespip == 1 {
		out[i] = &pb.DnstapAttribute{
			Id:    ah.attrsReqRespip[1].Id,
			Value: ah.attrsReqRespip[1].Value,
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
	out[i] = &pb.DnstapAttribute{Id: AttrNamePolicyAction, Value: actionConvDnstap[ah.action]}
	i++
	if len(ah.attrsReqRespip) > 0 {
		out[i] = &pb.DnstapAttribute{Id: AttrNameType, Value: TypeValueResponse}
	} else {
		out[i] = &pb.DnstapAttribute{Id: AttrNameType, Value: TypeValueQuery}
	}
	return out
}
