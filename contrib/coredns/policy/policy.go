// Package policy implements Policy Enforcement Point (PEP) for CoreDNS.
// It allows to connect CoreDNS to Policy Decision Point (PDP), send decision
// requests to the PDP and perform actions on DNS requests and replies according
// recieved decisions.
package policy

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin"

	"github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"

	"github.com/miekg/dns"

	"golang.org/x/net/context"
)

const (
	typeEDNS0Bytes = iota
	typeEDNS0Hex
	typeEDNS0IP
)

const (
	typeInvalid = iota
	typeRefuse
	typeAllow
	typeRedirect
	typeBlock
	typeLog

	actCount
)

var actionConv [actCount]string

func init() {
	actionConv[typeInvalid] = "invalid"
	actionConv[typeRefuse] = "refuse"
	actionConv[typeAllow] = "allow"
	actionConv[typeRedirect] = "redirect"
	actionConv[typeBlock] = "block"
	actionConv[typeLog] = "log"
}

var (
	errInvalidAction = errors.New("invalid action")
)

var stringToEDNS0MapType = map[string]uint16{
	"bytes":   typeEDNS0Bytes,
	"hex":     typeEDNS0Hex,
	"address": typeEDNS0IP,
}

type edns0Map struct {
	name     string
	dataType uint16
	destType string
	size     uint
	start    uint
	end      uint
}

// PolicyPlugin represents a plugin instance that can validate DNS
// requests and replies using PDP server.
type PolicyPlugin struct {
	Endpoints   []string
	Delay       uint
	Pending     uint
	options     map[uint16][]*edns0Map
	TapIO       dnstap.DnstapSender
	Next        plugin.Handler
	pdp         pep.Client
	DebugSuffix string
	ErrorFunc   func(dns.ResponseWriter, *dns.Msg, int) // failover error handler
}

const (
	debug = false
)

// Connect establishes connection to PDP server.
func (p *PolicyPlugin) Connect() error {
	log.Printf("[DEBUG] Endpoints: %v", p)
	if debug {
		p.pdp = newTestClientInit(
			&pdp.Response{Effect: pdp.PERMIT},
			&pdp.Response{Effect: pdp.PERMIT},
			nil,
			nil,
		)
	} else {
		p.pdp = pep.NewBalancedClient(p.Endpoints, p.Delay, p.Pending)
	}
	return p.pdp.Connect()
}

// Close terminates previously established connection.
func (p *PolicyPlugin) Close() {
	p.pdp.Close()
}

// AddEDNS0Map adds new EDNS0 to table.
func (p *PolicyPlugin) AddEDNS0Map(code, name, dataType, destType,
	sizeStr, startStr, endStr string) error {
	c, err := strconv.ParseUint(code, 0, 16)
	if err != nil {
		return fmt.Errorf("Could not parse EDNS0 code: %s", err)
	}
	size, err := strconv.ParseUint(sizeStr, 10, 32)
	if err != nil {
		return fmt.Errorf("Could not parse EDNS0 data size: %s", err)
	}
	start, err := strconv.ParseUint(startStr, 10, 32)
	if err != nil {
		return fmt.Errorf("Could not parse EDNS0 start index: %s", err)
	}
	end, err := strconv.ParseUint(endStr, 10, 32)
	if err != nil {
		return fmt.Errorf("Could not parse EDNS0 end index: %s", err)
	}
	if end <= start && end != 0 {
		return fmt.Errorf("End index should be > start index (actual %d <= %d)", end, start)
	}
	if end > size && size != 0 {
		return fmt.Errorf("End index should be <= size (actual %d > %d)", end, size)
	}
	ednsType, ok := stringToEDNS0MapType[dataType]
	if !ok {
		return fmt.Errorf("Invalid dataType for EDNS0 map: %s", dataType)
	}
	ecode := uint16(c)
	p.options[ecode] = append(p.options[ecode], &edns0Map{name, ednsType, destType, uint(size), uint(start), uint(end)})
	return nil
}

func resolve(status int) string {
	switch status {
	case -1:
		return "resolve:skip"
	case dns.RcodeSuccess:
		return "resolve:yes"
	case dns.RcodeServerFailure:
		return "resolve:failed"
	default:
		return "resolve:no"
	}
}

func join(key, value string) string { return "," + key + ":" + value }

func (p *PolicyPlugin) setDebugQueryAnswer(ah *attrHolder, r *dns.Msg, status int) {
	debugQueryInfo := resolve(status)

	var action string
	if len(ah.attrsRespRespip) > 0 {
		action = "pass"
	} else {
		action = actionConv[ah.action]
	}
	debugQueryInfo += join("query", action)
	for _, item := range ah.attrsRespDomain {
		debugQueryInfo += join(item.Id, item.Value)
	}
	if len(ah.attrsRespRespip) > 0 {
		debugQueryInfo += join("response", actionConv[ah.action])
		for _, item := range ah.attrsRespRespip {
			debugQueryInfo += join(item.Id, item.Value)
		}
	}

	hdr := dns.RR_Header{Name: r.Question[0].Name + p.DebugSuffix, Rrtype: dns.TypeTXT, Class: dns.ClassCHAOS, Ttl: 0}
	r.Answer = append(r.Answer, &dns.TXT{Hdr: hdr, Txt: []string{debugQueryInfo}})
}

func (p *PolicyPlugin) decodeDebugMsg(r *dns.Msg) bool {
	if r != nil && len(r.Question) > 0 {
		if r.Question[0].Qclass == dns.ClassCHAOS && r.Question[0].Qtype == dns.TypeTXT {
			if strings.HasSuffix(r.Question[0].Name, p.DebugSuffix) {
				r.Question[0].Name = strings.TrimSuffix(r.Question[0].Name, p.DebugSuffix)
				r.Question[0].Qtype = dns.TypeA
				r.Question[0].Qclass = dns.ClassINET
				return true
			}
		}
	}
	return false
}

func getNameType(r *dns.Msg) (string, uint16) {
	if r == nil || len(r.Question) == 0 {
		return ".", 0
	}
	return strings.ToLower(r.Question[0].Name), r.Question[0].Qtype
}

func getRemoteIP(w dns.ResponseWriter) string {
	addr := w.RemoteAddr().String()
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return ip
}

// ServeDNS implements the Handler interface.
func (p *PolicyPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	var (
		status    int = -1
		respMsg   *dns.Msg
		err       error
		sendExtra bool
	)

	debugQuery := p.decodeDebugMsg(r)
	ah := newAttrHolder(getNameType(r))
	ah.getAttrsFromEDNS0(getRemoteIP(w), r, p.options)

	// validate domain name (validation #1)
	err = p.validate(ah, "")
	if err != nil {
		status = dns.RcodeServerFailure
		goto Exit
	}

	if ah.action == typeAllow || ah.action == typeLog {
		// resolve domain name to IP
		responseWriter := new(Writer)
		status, err = plugin.NextOrFailure(p.Name(), p.Next, ctx, responseWriter, r)
		if err != nil {
			status = dns.RcodeServerFailure
		} else {
			respMsg = responseWriter.Msg
			address := extractRespIP(respMsg)
			// if external resolver ret code is not RcodeSuccess
			// address is not filled from the answer
			// in this case just pass through answer w/o validation
			if len(address) > 0 {
				// validate response IP (validation #2)
				err = p.validate(ah, address)
				if err != nil {
					status = dns.RcodeServerFailure
					goto Exit
				}
			}
		}
	}

	if debugQuery && (ah.action != typeRefuse) {
		p.setDebugQueryAnswer(ah, r, status)
		status = dns.RcodeSuccess
		err = nil
		goto Exit
	}

	if err != nil {
		goto Exit
	}

	switch ah.action {
	case typeAllow:
		r = respMsg
	case typeLog:
		sendExtra = true
		r = respMsg
	case typeRedirect:
		sendExtra = true
		status, err = p.redirect(ctx, r, ah.redirect)
	case typeBlock:
		sendExtra = true
		status = dns.RcodeNameError
	case typeRefuse:
		status = dns.RcodeRefused
	default:
		status = dns.RcodeServerFailure
		err = errInvalidAction
	}

Exit:
	r.Rcode = status
	r.Response = true
	if debugQuery {
		r.Question[0].Name = r.Question[0].Name + p.DebugSuffix
		r.Question[0].Qtype = dns.TypeTXT
		r.Question[0].Qclass = dns.ClassCHAOS
	}
	if p.TapIO != nil && status != dns.RcodeRefused && status != dns.RcodeServerFailure {
		if pw := dnstap.NewProxyWriter(w); pw != nil {
			pw.WriteMsg(r)
			var attrs []*dnstap.DnstapAttribute
			if sendExtra {
				attrs = ah.convertAttrs()
			}
			p.TapIO.SendCRExtraMsg(pw, attrs)
		}
	} else {
		w.WriteMsg(r)
	}
	return status, err
}

// Name implements the Handler interface
func (p *PolicyPlugin) Name() string { return "policy" }

func getNameClass(r *dns.Msg) (string, uint16) {
	if r == nil || len(r.Question) == 0 {
		return ".", 0
	}
	return r.Question[0].Name, r.Question[0].Qclass
}

func (p *PolicyPlugin) redirect(ctx context.Context, r *dns.Msg, dst string) (int, error) {
	var rr dns.RR
	cname := false

	ip := net.ParseIP(dst)
	qname, qclass := getNameClass(r)
	if ipv4 := ip.To4(); ipv4 != nil {
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeA, Class: qclass}
		rr.(*dns.A).A = ipv4
	} else if ipv6 := ip.To16(); ipv6 != nil {
		rr = new(dns.AAAA)
		rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeAAAA, Class: qclass}
		rr.(*dns.AAAA).AAAA = ipv6
	} else {
		dst = strings.TrimSuffix(dst, ".") + "."
		rr = new(dns.CNAME)
		rr.(*dns.CNAME).Hdr = dns.RR_Header{Name: qname, Rrtype: dns.TypeCNAME, Class: qclass}
		rr.(*dns.CNAME).Target = dst
		cname = true
	}

	if cname {
		origName := r.Question[0].Name
		r.Question[0].Name = dst
		responseWriter := new(Writer)
		status, err := plugin.NextOrFailure(p.Name(), p.Next, ctx, responseWriter, r)
		r.Question[0].Name = origName
		if err != nil {
			return dns.RcodeServerFailure, err
		}
		r.Answer = []dns.RR{rr}
		r.Answer = append(r.Answer, responseWriter.Msg.Answer...)
		r.Rcode = status
	} else {
		r.Answer = []dns.RR{rr}
		r.Rcode = dns.RcodeSuccess
	}

	return r.Rcode, nil
}

func (p *PolicyPlugin) validate(ah *attrHolder, addr string) error {
	req := &pdp.Request{}
	if len(addr) == 0 {
		req.Attributes = ah.attrsReqDomain
	} else {
		ah.makeReqRespip(addr)
		req.Attributes = ah.attrsReqRespip
	}

	response, err := p.pdp.Validate(req)
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s", err)
		return err
	}

	ah.addResponse(response, len(addr) > 0)
	return nil
}

func extractRespIP(m *dns.Msg) (address string) {
	if m == nil {
		return
	}
	for _, rr := range m.Answer {
		switch rr.Header().Rrtype {
		case dns.TypeA:
			if a, ok := rr.(*dns.A); ok {
				address = a.A.String()
			}
		case dns.TypeAAAA:
			if a, ok := rr.(*dns.AAAA); ok {
				address = a.AAAA.String()
			}
		}
	}
	return
}

type Writer struct {
	Msg *dns.Msg
}

func (w *Writer) Close() error                  { return nil }
func (w *Writer) TsigStatus() error             { return nil }
func (w *Writer) TsigTimersOnly(b bool)         { return }
func (w *Writer) Hijack()                       { return }
func (w *Writer) LocalAddr() (la net.Addr)      { return }
func (w *Writer) RemoteAddr() (ra net.Addr)     { return }
func (w *Writer) WriteMsg(m *dns.Msg) error     { w.Msg = m; return nil }
func (w *Writer) Write(buf []byte) (int, error) { return len(buf), nil }
