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
	sourceIP        *pdp.Attribute
	confAttrs       map[string]confAttrType
	attrsReqDomain  []*pdp.Attribute
	attrsRespDomain []*pdp.Attribute
	attrsReqRespip  []*pdp.Attribute
	attrsRespRespip []*pdp.Attribute
	action          byte
	redirect        string
	policyHit       bool
	attrsEdnsStart  int
	attrsTransfer   []*pdp.Attribute
}

func newAttrHolder(qName string, qType uint16, sourceIP string, conf map[string]confAttrType) *attrHolder {
	if conf == nil {
		conf = map[string]confAttrType{}
	}
	ret := &attrHolder{
		confAttrs:      conf,
		attrsReqDomain: make([]*pdp.Attribute, 4, 8),
		action:         typeInvalid,
		attrsEdnsStart: 4,
	}
	ret.attrsReqDomain[0] = &pdp.Attribute{Id: attrNameType, Type: "string", Value: typeValueQuery}
	ret.attrsReqDomain[1] = &pdp.Attribute{Id: attrNameDomainName, Type: "domain", Value: strings.TrimRight(qName, ".")}
	ret.attrsReqDomain[2] = &pdp.Attribute{Id: attrNameDNSQtype, Type: "string", Value: strconv.FormatUint(uint64(qType), 16)}
	ret.sourceIP = &pdp.Attribute{Id: attrNameSourceIP, Type: "address", Value: sourceIP}
	ret.attrsReqDomain[3] = ret.sourceIP
	return ret
}

func (ah *attrHolder) makeReqRespip(addr string) {
	ah.attrsReqRespip = []*pdp.Attribute{
		{Id: attrNameType, Type: "string", Value: typeValueResponse},
		{Id: attrNameAddress, Type: "address", Value: addr},
	}

	ah.attrsReqRespip = append(ah.attrsReqRespip, ah.attrsTransfer...)
}

func (ah *attrHolder) processConfAttr(attr *pdp.Attribute, t confAttrType) {
	switch t {
	case confAttrEdns:
		for _, item := range ah.attrsReqDomain[ah.attrsEdnsStart:] {
			if item.Id == attr.Id {
				return
			}
		}
		ah.attrsReqDomain = append(ah.attrsReqDomain, attr)
	case confAttrTransfer:
		ah.attrsTransfer = append(ah.attrsTransfer, attr)
	}
}

func (ah *attrHolder) addResponse(r *pdp.Response, respip bool) {
	i := 0
	l := len(r.Obligation)
	switch r.Effect {
	case pdp.Response_PERMIT:
		// don't overwrite "Log" action from previous validation
		if ah.action != typeLog {
			ah.action = typeAllow
		}
		for i < l {
			item := r.Obligation[i]
			if item.Id == attrNameLog {
				ah.action = typeLog
			} else if !respip {
				if t, ok := ah.confAttrs[item.Id]; ok {
					ah.processConfAttr(item, t)
					if t == confAttrEdns {
						//remove duplicate ends attr from response
						l--
						r.Obligation[i] = r.Obligation[l]
						continue
					}
				}
			}
			i++
		}

	case pdp.Response_DENY:
		ah.action = typeBlock
		for i < l {
			item := r.Obligation[i]
			switch item.Id {
			case attrNameRefuse:
				ah.action = typeRefuse
			case attrNameRedirectTo:
				ah.action = typeRedirect
				ah.redirect = item.Value
			case attrNameDrop:
				ah.action = typeDrop
			default:
				if !respip {
					if t, ok := ah.confAttrs[item.Id]; ok && t == confAttrEdns {
						ah.processConfAttr(item, t)
						//remove duplicate ends attr from response
						l--
						r.Obligation[i] = r.Obligation[l]
						continue
					}
				}
			}
			i++
		}
	default:
		log.Printf("[ERROR] PDP Effect: %s", r.Effect)
		ah.action = typeInvalid
	}

	if !respip {
		ah.attrsRespDomain = r.Obligation[:l]
	} else {
		ah.attrsRespRespip = r.Obligation[:l]
	}
	return
}

func (ah *attrHolder) convertAttrs() []*pb.DnstapAttribute {
	if !ah.policyHit {
		return ah.convertAttrsEdns()
	}
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
	out[i] = &pb.DnstapAttribute{Id: attrNamePolicyAction, Value: actionConvDnstap[ah.action]}
	i++
	if len(ah.attrsReqRespip) > 0 {
		out[i] = &pb.DnstapAttribute{Id: attrNameType, Value: typeValueResponse}
	} else {
		out[i] = &pb.DnstapAttribute{Id: attrNameType, Value: typeValueQuery}
	}
	return out
}

func (ah *attrHolder) convertAttrsEdns() []*pb.DnstapAttribute {
	attrsEdns := ah.attrsReqDomain[ah.attrsEdnsStart:]
	out := make([]*pb.DnstapAttribute, len(attrsEdns))
	for i, item := range attrsEdns {
		out[i] = &pb.DnstapAttribute{
			Id:    item.Id,
			Value: item.Value,
		}
	}
	return out
}
