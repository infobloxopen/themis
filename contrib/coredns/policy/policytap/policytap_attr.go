package policytap

import (
	"net"

	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/miekg/dns"
)

// PdpAttr2DnstapAttr converts service.Attribute to DnstapAttribute
func PdpAttr2DnstapAttr(a *pdp.Attribute) *DnstapAttribute {
	return &DnstapAttribute{Id: &a.Id, Type: &a.Type, Value: &a.Value}
}

// ConvertAttrs converts slice of service.Attribute to slice of DnstapAttribute
func ConvertAttrs(in []*pdp.Attribute) []*DnstapAttribute {
	out := make([]*DnstapAttribute, 0, len(in))
	for _, a := range in {
		out = append(out, PdpAttr2DnstapAttr(a))
	}
	return out
}

// AddDnstapAttrs adds DnstapAttribute to PolicyHitMessage
func (phm *PolicyHitMessage) AddDnstapAttrs(attrs []*DnstapAttribute) {
	phm.Attributes = append(phm.Attributes, attrs...)
}

// AddPdpAttrs adds service.Attribute to PolicyHitMessage
func (phm *PolicyHitMessage) AddPdpAttrs(attrs []*pdp.Attribute) {
	for _, a := range attrs {
		if h, ok := specialAttrs[a.Id]; ok {
			h(phm, a)
		} else {
			phm.Attributes = append(phm.Attributes, PdpAttr2DnstapAttr(a))
		}
	}
}

type attrHandler func(*PolicyHitMessage, *pdp.Attribute)

var specialAttrs = map[string]attrHandler{
	"policy_id":   policyIDHandler,
	"refuse":      refuseHandler,
	"redirect_to": redirectHandler,
}

func policyIDHandler(phm *PolicyHitMessage, a *pdp.Attribute) {
	phm.PolicyId = []byte(a.Value)
}

func refuseHandler(phm *PolicyHitMessage, a *pdp.Attribute) {
	if a.Value != "" &&
		*phm.PolicyAction != PolicyHitMessage_POLICY_ACTION_PASSTHROUGH &&
		*phm.PolicyAction != PolicyHitMessage_POLICY_ACTION_REDIRECT {

		act := PolicyHitMessage_POLICY_ACTION_REFUSE
		phm.PolicyAction = &act
	}
}

func redirectHandler(phm *PolicyHitMessage, a *pdp.Attribute) {
	if a.Value != "" {
		if *phm.PolicyAction != PolicyHitMessage_POLICY_ACTION_PASSTHROUGH {
			act := PolicyHitMessage_POLICY_ACTION_REDIRECT
			phm.PolicyAction = &act
		}

		rType := uint32(dns.TypeCNAME)
		if ip := net.ParseIP(a.Value); ip != nil {
			if ip.To4() != nil {
				rType = uint32(dns.TypeA)
			} else {
				rType = uint32(dns.TypeAAAA)
			}
		}

		phm.RedirectRrType = &rType
		phm.RedirectRrData = &a.Value
	}
}
