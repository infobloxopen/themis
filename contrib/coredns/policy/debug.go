package policy

import (
	"bytes"
	"strings"

	"github.com/infobloxopen/themis/pdp"
	"github.com/miekg/dns"
)

const txtLenLimit = 255
const initTxtCount = 5

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
	orgName   string
	msgBounds []int
}

func newDbgMessenger(name, id string) *dbgMessenger {
	dm := &dbgMessenger{orgName: name, msgBounds: make([]int, 0, initTxtCount)}
	if id != "" {
		dm.appendDebugID(id)
	}
	return dm
}

func (p *policyPlugin) patchDebugMsg(r *dns.Msg) *dbgMessenger {
	if r == nil || len(r.Question) <= 0 {
		return nil
	}

	q := &r.Question[0]
	if q.Qclass != dns.ClassCHAOS || q.Qtype != dns.TypeTXT || !strings.HasSuffix(q.Name, p.conf.debugSuffix) {
		return nil
	}

	orgName := q.Name
	q.Name = dns.Fqdn(strings.TrimSuffix(q.Name, p.conf.debugSuffix))
	q.Qtype = dns.TypeA
	q.Qclass = dns.ClassINET

	return newDbgMessenger(orgName, p.conf.debugID)
}

func (dm *dbgMessenger) restoreDebugMsg(r *dns.Msg) {
	if len(r.Question) > 0 {
		q := &r.Question[0]
		q.Name = dm.orgName
		q.Qtype = dns.TypeTXT
		q.Qclass = dns.ClassCHAOS
	}
}

func (dm *dbgMessenger) setDebugQueryPassthroughAnswer(r *dns.Msg) {
	dm.appendPassthrough()
	r.Answer = dm.makeDebugQueryAnswerRR()
	r.Ns = nil
}

func (dm *dbgMessenger) setDebugQueryAnswer(r *dns.Msg, status int) {
	dm.appendResolution(status)
	r.Answer = dm.makeDebugQueryAnswerRR()
}

func (dm *dbgMessenger) makeDebugQueryAnswerRR() []dns.RR {
	return []dns.RR{
		&dns.TXT{
			Hdr: dns.RR_Header{
				Name:   dm.orgName,
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassCHAOS,
			},
			Txt: dm.txtMsgs(),
		},
	}
}

func (dm *dbgMessenger) txtMsgs() []string {
	msgs := dm.String()
	txts := make([]string, 0, initTxtCount)
	s := 0
	for _, end := range dm.msgBounds {
		for s < end {
			e := end
			if s+txtLenLimit < end {
				e = s + txtLenLimit
			}
			txts = append(txts, msgs[s:e])
			s = e
		}
	}
	return txts
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
	for i, a := range attrs {
		if i > 0 {
			dm.WriteString(", ")
		}
		n, v := debugNameVal(a)
		dm.WriteString(n)
		dm.WriteString(": ")
		dm.WriteString(v)
	}
	dm.WriteString("]")
}

func (dm *dbgMessenger) appendDebugID(id string) {
	dm.WriteString("Ident: ")
	dm.WriteString(id)
	dm.msgBounds = append(dm.msgBounds, dm.Len())
}

func (dm *dbgMessenger) appendPassthrough() {
	dm.WriteString("Passthrough: yes")
	dm.msgBounds = append(dm.msgBounds, dm.Len())
}

func (dm *dbgMessenger) appendResolution(rCode int) {
	dm.WriteString("Domain resolution: ")
	switch rCode {
	case -1:
		dm.WriteString("skipped")
	case dns.RcodeSuccess:
		dm.WriteString("resolved")
	case dns.RcodeServerFailure:
		dm.WriteString("failed")
	default:
		dm.WriteString("not resolved")
	}
	dm.msgBounds = append(dm.msgBounds, dm.Len())
}

func (dm *dbgMessenger) appendDefaultDecision(attrs []pdp.AttributeAssignment) {
	dm.appendAttrs("Default decision", attrs)
	dm.msgBounds = append(dm.msgBounds, dm.Len())
}

func (dm *dbgMessenger) appendResponse(res *pdp.Response) {
	dm.WriteString("PDP response {Effect: ")
	dm.WriteString(pdp.EffectNameFromEnum(res.Effect))
	dm.WriteString(", ")
	dm.appendAttrs("Obligations", res.Obligations)
	dm.WriteString("}")
	dm.msgBounds = append(dm.msgBounds, dm.Len())
}
