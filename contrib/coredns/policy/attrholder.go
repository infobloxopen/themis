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
	attrsRequest  []*pdp.Attribute
	attrsResponse []*pdp.Attribute
	action        byte
	redirect      string
}

func newAttrHolder(qName string, qType uint16) *attrHolder {
	ret := &attrHolder{
		attrsRequest: make([]*pdp.Attribute, 2, 8),
		action:       typeInvalid,
	}
	ret.attrsRequest[0] = &pdp.Attribute{Id: AttrNameDomainName, Type: "domain", Value: strings.TrimRight(qName, ".")}
	ret.attrsRequest[1] = &pdp.Attribute{Id: AttrNameDNSQtype, Type: "string", Value: strconv.FormatUint(uint64(qType), 16)}
	return ret
}

func (ah *attrHolder) addRespip(addr string) {
	ah.attrsRequest = append(ah.attrsRequest, &pdp.Attribute{Id: AttrNameAddress, Type: "address", Value: addr})
}

func (ah *attrHolder) addResponse(r *pdp.Response) {
	ah.attrsResponse = r.Obligation

	switch r.Effect {
	case pdp.Response_PERMIT:
		for _, item := range r.Obligation {
			if item.Id == AttrNameLog {
				ah.action = typeLog
				return
			}
		}
		ah.action = typeAllow
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
	size := len(ah.attrsRequest) + len(ah.attrsResponse) + 1
	out := make([]*pb.DnstapAttribute, size)
	i := 0
	for _, item := range ah.attrsRequest {
		out[i] = &pb.DnstapAttribute{
			Id:    item.Id,
			Value: item.Value,
		}
		i++
	}
	for _, item := range ah.attrsResponse {
		out[i] = &pb.DnstapAttribute{
			Id:    item.Id,
			Value: item.Value,
		}
		i++
	}
	out[i] = &pb.DnstapAttribute{Id: AttrNamePolicyAction, Value: actionConvDnstap[ah.action]}
	return out
}
