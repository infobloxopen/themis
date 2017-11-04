// Package policy implements Policy Enforcement Point (PEP) for CoreDNS.
// It allows to connect CoreDNS to Policy Decision Point (PDP), send decision
// requests to the PDP and perform actions on DNS requests and replies according
// recieved decisions.
package policy

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin"

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
	TapIO       DnstapSender
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

func parseOptionGroup(ah *attrHolder, data []byte, options []*edns0Map) bool {
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

func (p *PolicyPlugin) getAttrsFromEDNS0(ah *attrHolder, r *dns.Msg, ip string) {
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
		options, ok := p.options[optLocal.Code]
		if !ok {
			continue
		}
		srcIpFound := parseOptionGroup(ah, optLocal.Data, options)
		if srcIpFound {
			ipId = "proxy_source_ip"
		}
	}

Exit:
	ah.attrsReqDomain = append(ah.attrsReqDomain, &pdp.Attribute{Id: ipId, Type: "address", Value: ip})
	return
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

func getNameClass(r *dns.Msg) (string, uint16) {
	if r == nil || len(r.Question) == 0 {
		return ".", 0
	}
	return r.Question[0].Name, r.Question[0].Qclass
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
	p.getAttrsFromEDNS0(ah, r, getRemoteIP(w))

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
		sendExtra = false
	}
	if p.TapIO != nil && status != dns.RcodeRefused && status != dns.RcodeServerFailure {
		if pw := NewProxyWriter(w); pw != nil {
			pw.WriteMsg(r)
			if !sendExtra {
				ah = nil
			}
			p.TapIO.SendCRExtraMsg(pw, ah)
		}
	} else {
		w.WriteMsg(r)
	}
	return status, err
}

// Name implements the Handler interface
func (p *PolicyPlugin) Name() string { return "policy" }

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
