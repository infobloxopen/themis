package policy

import (
	"encoding/hex"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/miekg/dns"
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

func parseHex(data []byte, option *edns0Map) string {
	size := uint(len(data))
	// if option.size == 0 - don't check size
	if option.size > 0 {
		if size != option.size {
			// skip parsing option with wrong size
			return ""
		}
	}
	start := uint(0)
	if option.start < size {
		// set start index
		start = option.start
	} else {
		// skip parsing option if start >= data size
		return ""
	}
	end := size
	// if option.end == 0 - return data[start:]
	if option.end > 0 {
		if option.end <= size {
			// set end index
			end = option.end
		} else {
			// skip parsing option if end > data size
			return ""
		}
	}
	return hex.EncodeToString(data[start:end])
}

func (ah *attrHolder) parseOptionGroup(data []byte, options []*edns0Map) bool {
	srcIpFound := false
	for _, option := range options {
		var value string
		switch option.dataType {
		case typeEDNS0Bytes:
			value = string(data)
		case typeEDNS0Hex:
			value = parseHex(data, option)
			if value == "" {
				continue
			}
		case typeEDNS0IP:
			ip := net.IP(data)
			value = ip.String()
		}
		if option.name == "source_ip" {
			srcIpFound = true
		}
		ah.attrsReqDomain = append(ah.attrsReqDomain, &pdp.Attribute{Id: option.name, Type: option.destType, Value: value})
	}
	return srcIpFound
}

func (ah *attrHolder) getAttrsFromEDNS0(ip string, r *dns.Msg, options map[uint16][]*edns0Map) {
	ipId := "source_ip"

	o := r.IsEdns0()
	if o == nil {
		goto Exit
	}

	for _, opt := range o.Option {
		optLocal, local := opt.(*dns.EDNS0_LOCAL)
		if !local {
			continue
		}
		options, ok := options[optLocal.Code]
		if !ok {
			continue
		}
		srcIpFound := ah.parseOptionGroup(optLocal.Data, options)
		if srcIpFound {
			ipId = "proxy_source_ip"
		}
	}

Exit:
	ah.attrsReqDomain = append(ah.attrsReqDomain, &pdp.Attribute{Id: ipId, Type: "address", Value: ip})
	return
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
