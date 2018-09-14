package policy

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/miekg/dns"

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
		conf:            newConfig(),
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
		resolveFailed bool
	)
	p.wg.Add(1)
	defer p.wg.Done()

	dbgMsgr := p.patchDebugMsg(r)

	// turn off default Cq and Cr dnstap messages
	resetCqCr(ctx)

	attrBuffer := p.attrPool.Get()
	tmpAttrBuffer := p.attrPool.Get()

	ah := newAttrHolder(attrBuffer, p.conf.attrs)
	dn := ah.addDnsQuery(w, r, p.conf.options)

	defer func() {
		if p.conf.attrs.hasMetrics() {
			metricsList := ah.attrList(tmpAttrBuffer, attrListTypeMetrics)
			for _, a := range metricsList {
				p.attrGauges.Inc(a)
			}
		}

		if r != nil {
			r.Rcode = status
			r.Response = true
			clearECS(r)

			if dbgMsgr != nil {
				dbgMsgr.restoreDebugMsg(r)
			}

			if ah.actionValue() != actionDrop && (status != dns.RcodeServerFailure || resolveFailed) {
				w.WriteMsg(r)
			}

			if p.tapIO != nil && dbgMsgr == nil {
				p.tapIO.sendCRExtraMsg(w, r, ah.dnstapList())
			}
		}

		p.attrPool.Put(tmpAttrBuffer)
		p.attrPool.Put(attrBuffer)
	}()

	for _, s := range p.conf.passthrough {
		if strings.HasSuffix(dn, s) {
			nw := nonwriter.New(w)
			_, err := plugin.NextOrFailure(p.Name(), p.next, ctx, nw, r)
			r = nw.Msg
			if r != nil {
				status = r.Rcode

				if dbgMsgr != nil {
					dbgMsgr.setDebugQueryPassthroughAnswer(r)
					status = dns.RcodeSuccess
				}
			}

			return dns.RcodeSuccess, err
		}
	}

	// validate domain name (validation #1)
	goon, err := p.validate(tmpAttrBuffer, ah, attrListTypeVal1, dbgMsgr)
	if err != nil {
		status = dns.RcodeServerFailure
		return dns.RcodeSuccess, err
	}

	if goon {
		// resolve domain name to IP
		nw := nonwriter.New(w)
		_, err := plugin.NextOrFailure(p.Name(), p.next, ctx, nw, r)
		if err != nil {
			resolveFailed = true
			status = dns.RcodeServerFailure

			if dbgMsgr != nil {
				dbgMsgr.setDebugQueryAnswer(r, status)
				status = dns.RcodeSuccess
				return dns.RcodeSuccess, nil
			}

			return dns.RcodeSuccess, err
		}

		respMsg = nw.Msg
		if respMsg == nil {
			r = nil
			return dns.RcodeSuccess, nil
		}

		status = respMsg.Rcode
		if status == dns.RcodeServerFailure {
			resolveFailed = true
		}

		address := getRespIP(respMsg)
		// if external resolver ret code is not RcodeSuccess
		// address is not filled from the answer
		// in this case just pass through answer w/o validation
		if address != nil {
			ah.addAddressAttr(address)

			// validate response IP (validation #2)
			if _, err := p.validate(tmpAttrBuffer, ah, attrListTypeVal2, dbgMsgr); err != nil {
				status = dns.RcodeServerFailure
				return dns.RcodeSuccess, err
			}
		}
	}

	if dbgMsgr != nil && ah.actionValue() != actionRefuse {
		dbgMsgr.setDebugQueryAnswer(r, status)
		status = dns.RcodeSuccess
		return dns.RcodeSuccess, nil
	}

	switch ah.actionValue() {
	case actionAllow:
		r = respMsg
		if ah.logValue() > 0 {
			r = resetTTL(respMsg)
		}
		return dns.RcodeSuccess, nil

	case actionRedirect:
		var err error
		status, err = p.setRedirectQueryAnswer(ctx, w, r, ah.redirectValue())
		r.AuthenticatedData = false
		return dns.RcodeSuccess, err

	case actionBlock:
		status = dns.RcodeNameError
		r.AuthenticatedData = false
		return dns.RcodeSuccess, nil

	case actionRefuse:
		status = dns.RcodeRefused
		return dns.RcodeSuccess, nil

	case actionDrop:
		return dns.RcodeSuccess, nil
	}

	status = dns.RcodeServerFailure
	return dns.RcodeSuccess, errInvalidAction
}
