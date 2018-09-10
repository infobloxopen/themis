package policy

import (
	"fmt"
	"net"
	"strconv"

	"github.com/infobloxopen/go-trees/domain"
	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	"github.com/miekg/dns"

	"github.com/infobloxopen/themis/pdp"
)

var emptyCtx, _ = pdp.NewContext(nil, 0, nil)
var emptyAttr = pdp.MakeAttributeAssignment(pdp.Attribute{}, pdp.UndefinedValue)

// attrHolder holds all attributes related to particular DNS query
type attrHolder struct {
	attrs []pdp.AttributeAssignment
	cfg   *attrsConfig
}

// newAttrHolder constructs attrHolder object which holds all attributes related to
// particular DNS query
// buf is a slice with underlying allocated buffer for attributes. If buf is nil
// or capacity is not enough to store all predefined+configured attributes then new
// buffer is allocated
// conf must have mappings for all predefined and configured attributes
func newAttrHolder(buf []pdp.AttributeAssignment, conf *attrsConfig) *attrHolder {
	attrCnt := len(conf.attrInds)
	if cap(buf) < attrCnt {
		buf = make([]pdp.AttributeAssignment, attrCnt)
	}
	ah := &attrHolder{attrs: buf[:attrCnt], cfg: conf}
	for i := range ah.attrs {
		ah.attrs[i] = emptyAttr
	}
	return ah
}

// addDnsQuery adds some predefined attributes from DNS query as well as configured
// attributes from EDNS record
func (ah *attrHolder) addDnsQuery(w dns.ResponseWriter, r *dns.Msg, optMap map[uint16][]*edns0Opt) string {
	qName, qType := getNameAndType(r)
	dn, err := domain.MakeNameFromString(qName)
	if err != nil {
		panic(fmt.Errorf("Can't treat %q as domain name: %s", qName, err))
	}
	ah.attrs[attrIndexDomainName] = pdp.MakeDomainAssignment(attrNameDomainName, dn)
	ah.attrs[attrIndexDNSQtype] = pdp.MakeStringAssignment(attrNameDNSQtype, strconv.FormatUint(uint64(qType), 16))

	if srcIP := getRemoteIP(w); srcIP != nil {
		ah.attrs[attrIndexSourceIP] = pdp.MakeAddressAssignment(attrNameSourceIP, srcIP)
	}

	extractOptionsFromEDNS0(r, optMap, func(b []byte, opts []*edns0Opt) {
		for _, o := range opts {
			if a, ok := makeAssignmentByType(o, b); ok {
				ah.attrs[o.attrInd] = a
			}
		}
	})
	return qName
}

func makeAssignmentByType(o *edns0Opt, b []byte) (pdp.AttributeAssignment, bool) {
	switch o.dataType {
	case typeEDNS0Bytes:
		return pdp.MakeStringAssignment(o.name, string(b)), true

	case typeEDNS0Hex:
		s := o.makeHexString(b)
		return pdp.MakeStringAssignment(o.name, s), s != ""

	case typeEDNS0IP:
		return pdp.MakeAddressAssignment(o.name, net.IP(b)), true
	}

	panic(fmt.Errorf("unknown attribute type %d", o.dataType))
}

// addAddressAttr adds predefined Address attribute
func (ah *attrHolder) addAddressAttr(ip net.IP) {
	ah.attrs[attrIndexAddress] = pdp.MakeAddressAssignment(attrNameAddress, ip)
}

// addAttrList adds or replaces attributes from the attrs list
func (ah *attrHolder) addAttrList(attrs []pdp.AttributeAssignment) {
	for _, a := range attrs {
		if ind, ok := ah.cfg.attrInds[a.GetID()]; ok {
			ah.attrs[ind] = a
		}
	}
}

// resetAttrList resets values of configured attribute list to preconfigured values
func (ah *attrHolder) resetAttrList(listType int) {
	for _, ac := range ah.cfg.confLists[listType] {
		if ac.value != pdp.UndefinedValue {
			ah.attrs[ac.index] = pdp.MakeExpressionAssignment(ac.name, ac.value)
		}
	}
}

// attrList returns non-empty attributes belonging to particular list
// buf is a slice to use as a preallocated buffer. If capacity is not enough to hold
// all returned attributes then the buffer will be automatically reallocated
func (ah *attrHolder) attrList(buf []pdp.AttributeAssignment, listType int) []pdp.AttributeAssignment {
	ah.resetAttrList(listType)

	list := buf[:0]
	for _, ac := range ah.cfg.confLists[listType] {
		attr := ah.attrs[ac.index]
		if attr.GetID() == "" {
			continue
		}
		list = append(list, attr)
	}
	return list
}

// dnstapList returns list of non-empty dnstap attributes according to defined log level
func (ah *attrHolder) dnstapList() []*pb.DnstapAttribute {
	cfgList := ah.cfg.confLists[attrListTypeDnstap+ah.logValue()]
	if len(cfgList) <= 0 {
		return nil
	}

	ah.resetAttrList(attrListTypeDnstap + ah.logValue())
	out := make([]*pb.DnstapAttribute, 0, len(cfgList))

	for _, ac := range cfgList {
		attr := ah.attrs[ac.index]
		if attr.GetID() == "" {
			continue
		}
		out = append(out, newDnstapAttribute(attr))
	}
	return out
}

// redirectValue returns the value of Redirect attribute
func (ah *attrHolder) redirectValue() string {
	if rdr, err := ah.attrs[attrIndexRedirectTo].GetString(emptyCtx); err == nil {
		return rdr
	}
	return ""
}

// actionValue returns the Action
func (ah *attrHolder) actionValue() int {
	act64, err := ah.attrs[attrIndexPolicyAction].GetInteger(emptyCtx)
	if err == nil && act64 >= 0 && act64 < actionsTotal {
		return int(act64)
	}
	return actionInvalid
}

// logValue returns the dnstap log level
func (ah *attrHolder) logValue() int {
	log64, err := ah.attrs[attrIndexLog].GetInteger(emptyCtx)
	if err == nil && log64 > 0 && log64 < maxDnstapLists {
		return int(log64)
	}
	return 0
}

// resetAttribute resets attribute to empty value
func (ah *attrHolder) resetAttribute(ind int) {
	ah.attrs[ind] = emptyAttr
}
