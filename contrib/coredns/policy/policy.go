package policy

import (
	"errors"
	"strings"
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/miekg/dns"
	"golang.org/x/net/context"

	"github.com/infobloxopen/themis/pep"
)

var errInvalidAction = errors.New("invalid action")

// policyPlugin represents a plugin instance that can validate DNS
// requests and replies using PDP server.
type policyPlugin struct {
	conf            config
	tapIO           dnstapSender
	trace           plugin.Handler
	next            plugin.Handler
	pdp             pep.Client
	attrPool        attrPool
	attrGauges      *AttrGauge
	connAttempts    map[string]*uint32
	unkConnAttempts *uint32
	wg              sync.WaitGroup
}

func newPolicyPlugin() *policyPlugin {
	return &policyPlugin{
		conf: config{
			options:     make(map[uint16][]*edns0Opt),
			custAttrs:   make(map[string]custAttr),
			connTimeout: -1,
			maxReqSize:  -1,
			maxResAttrs: 64,
		},
		connAttempts:    make(map[string]*uint32),
		unkConnAttempts: new(uint32),
	}
}

// Name implements the Handler interface
func (p *policyPlugin) Name() string { return "policy" }

// ServeDNS implements the Handler interface.
func (p *policyPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	var (
		status        = -1
		respMsg       *dns.Msg
		err           error
		resolveFailed bool
	)
	p.wg.Add(1)
	defer p.wg.Done()

	// turn off default Cq and Cr dnstap messages
	resetCqCr(ctx)

	debugQuery := p.patchDebugMsg(r)

	ah := newAttrHolderWithDnReq(w, r, p.conf.options, p.attrGauges)

	for _, s := range p.conf.passthrough {
		if strings.HasSuffix(ah.dn, s) {
			responseWriter := nonwriter.New(w)
			_, err = plugin.NextOrFailure(p.Name(), p.next, ctx, responseWriter, r)
			r = responseWriter.Msg
			status = r.Rcode
			if debugQuery {
				hdr := dns.RR_Header{Name: r.Question[0].Name + p.conf.debugSuffix,
					Rrtype: dns.TypeTXT, Class: dns.ClassCHAOS, Ttl: 0}
				r.Answer = []dns.RR{&dns.TXT{Hdr: hdr, Txt: []string{"action:passthrough"}}}
				r.Ns = nil
				status = dns.RcodeSuccess
			}
			goto Exit
		}
	}

	// validate domain name (validation #1)
	err = p.validate(ah)
	if err != nil {
		status = dns.RcodeServerFailure
		goto Exit
	}

	if ah.action == actionAllow || ah.action == actionLog {
		// resolve domain name to IP
		responseWriter := nonwriter.New(w)
		_, err = plugin.NextOrFailure(p.Name(), p.next, ctx, responseWriter, r)
		if err != nil {
			resolveFailed = true
			status = dns.RcodeServerFailure
		} else {
			respMsg = responseWriter.Msg
			status = respMsg.Rcode
			if status == dns.RcodeServerFailure {
				resolveFailed = true
			}
			address := getRespIP(respMsg)
			// if external resolver ret code is not RcodeSuccess
			// address is not filled from the answer
			// in this case just pass through answer w/o validation
			if len(address) > 0 {
				ah.addIPReq(address)
				// validate response IP (validation #2)
				err = p.validate(ah)
				if err != nil {
					status = dns.RcodeServerFailure
					goto Exit
				}
			}
		}
	}

	if debugQuery && (ah.action != actionRefuse) {
		p.setDebugQueryAnswer(ah, r, status)
		status = dns.RcodeSuccess
		err = nil
		goto Exit
	}

	if err != nil {
		goto Exit
	}

	switch ah.action {
	case actionAllow:
		r = respMsg
	case actionLog:
		r = respMsg
	case actionRedirect:
		status, err = p.setRedirectQueryAnswer(ctx, w, r, ah.dst)
		r.AuthenticatedData = false
	case actionBlock:
		status = dns.RcodeNameError
		r.AuthenticatedData = false
	case actionRefuse:
		status = dns.RcodeRefused
	case actionDrop:
		return dns.RcodeSuccess, nil
	default:
		status = dns.RcodeServerFailure
		err = errInvalidAction
	}

Exit:
	r.Rcode = status
	r.Response = true
	clearECS(r)
	if debugQuery {
		r.Question[0].Name = r.Question[0].Name + p.conf.debugSuffix
		r.Question[0].Qtype = dns.TypeTXT
		r.Question[0].Qclass = dns.ClassCHAOS
	}

	if status != dns.RcodeServerFailure || resolveFailed {
		w.WriteMsg(r)
	}

	if p.tapIO != nil && !debugQuery {
		p.tapIO.sendCRExtraMsg(w, r, ah)
	}
	return dns.RcodeSuccess, err
}
