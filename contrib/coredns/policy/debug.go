package policy

import (
	"bytes"
	"strings"

	"github.com/infobloxopen/themis/pdp"
	"github.com/miekg/dns"
)

var actionNames [actionsTotal]string

func init() {
	actionNames[actionInvalid] = "invalid"
	actionNames[actionDrop] = "drop"
	actionNames[actionAllow] = "allow"
	actionNames[actionBlock] = "block"
	actionNames[actionRedirect] = "redirect"
	actionNames[actionRefuse] = "refuse"
}

type dbgMessenger struct {
	bytes.Buffer
	suffix string
}

func newDbgMessenger(suff, id string) *dbgMessenger {
	dm := &dbgMessenger{suffix: suff}
	if id != "" {
		dm.appendDebugID(id)
	}
	return dm
}

func (p *policyPlugin) patchDebugMsg(r *dns.Msg) *dbgMessenger {
	if r == nil || len(r.Question) <= 0 {
		return nil
	}

	q := r.Question[0]
	if q.Qclass != dns.ClassCHAOS || q.Qtype != dns.TypeTXT || !strings.HasSuffix(q.Name, p.conf.debugSuffix) {
		return nil
	}

	q.Name = dns.Fqdn(strings.TrimSuffix(q.Name, p.conf.debugSuffix))
	q.Qtype = dns.TypeA
	q.Qclass = dns.ClassINET

	r.Question[0] = q

	return newDbgMessenger(p.conf.debugSuffix, p.conf.debugID)
}

func (dm *dbgMessenger) setDebugQueryPassthroughAnswer(dn string, r *dns.Msg) {
	dm.appendPassthrough()
	r.Answer = dm.makeDebugQueryAnswerRR(dn)
	r.Ns = nil
}

func (dm *dbgMessenger) setDebugQueryAnswer(dn string, r *dns.Msg, status int) {
	dm.appendResolution(status)
	r.Answer = dm.makeDebugQueryAnswerRR(dn)
}

func (dm *dbgMessenger) makeDebugQueryAnswerRR(dn string) []dns.RR {
	return []dns.RR{
		&dns.TXT{
			Hdr: dns.RR_Header{
				Name:   dn + dm.suffix,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassCHAOS,
			},
			Txt: []string{dm.String()},
		},
	}
}

func debugNameVal(a pdp.AttributeAssignment) (n, v string) {
	if a.GetID() == attrNamePolicyAction {
		iv, err := a.GetInteger(emptyCtx)
		if err == nil && iv >= 0 && iv < actionsTotal {
			n = attrNamePolicyAction
			v = actionNames[iv]
			return
		}
	}
	n, _, v, _ = a.Serialize(emptyCtx)
	return
}

func (dm *dbgMessenger) appendAttrs(msg string, attrs []pdp.AttributeAssignment) {
	dm.WriteString(msg)
	dm.WriteString(": [")
	for _, a := range attrs {
		n, v := debugNameVal(a)
		dm.WriteString(n)
		dm.WriteString(": ")
		dm.WriteString(v)
		dm.WriteString(", ")
	}
	dm.WriteString("], ")
}

func (dm *dbgMessenger) appendDebugID(id string) {
	dm.WriteString("Ident: ")
	dm.WriteString(id)
	dm.WriteString(", ")
}

func (dm *dbgMessenger) appendPassthrough() {
	dm.WriteString("Passthrough: yes, ")
}

func (dm *dbgMessenger) appendResolution(rCode int) {
	dm.WriteString("Domain resolution: ")
	switch rCode {
	case -1:
		dm.WriteString("skipped, ")
	case dns.RcodeSuccess:
		dm.WriteString("resolved, ")
	case dns.RcodeServerFailure:
		dm.WriteString("failed, ")
	default:
		dm.WriteString("not resolved, ")
	}
}

func (dm *dbgMessenger) appendResponse(res *pdp.Response) {
	dm.WriteString("PDP response {Effect: ")
	dm.WriteString(pdp.EffectNameFromEnum(res.Effect))
	dm.WriteString(", ")
	dm.appendAttrs("Obligations", res.Obligations)
	dm.WriteString("}, ")
}
