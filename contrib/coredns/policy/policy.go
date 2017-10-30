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
	"github.com/coredns/coredns/plugin/pkg/trace"
	"github.com/coredns/coredns/request"

	"github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
	pb "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pep"

	"github.com/miekg/dns"

	ot "github.com/opentracing/opentracing-go"

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

	actCount
)

const (
	resolveYes = iota
	resolveNo
	resolveFailed
	resolveSkip

	resCount
)

var resConv [resCount]string

func init() {
	resConv[resolveYes] = "resolve: yes"
	resConv[resolveNo] = "resolve: no"
	resConv[resolveFailed] = "resolve: failed"
	resConv[resolveSkip] = "resolve: skip"
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
	options     map[uint16][]*edns0Map
	TapIO       dnstap.DnstapSender
	Trace       plugin.Handler
	Next        plugin.Handler
	pdp         pep.Client
	DebugSuffix string
	ErrorFunc   func(dns.ResponseWriter, *dns.Msg, int) // failover error handler
}

// Connect establishes connection to PDP server.
func (p *PolicyPlugin) Connect() error {
	log.Printf("[DEBUG] Connecting %v", p)
	var tracer ot.Tracer
	if p.Trace != nil {
		if t, ok := p.Trace.(trace.Trace); ok {
			tracer = t.Tracer()
		}
	}
	p.pdp = pep.NewBalancedClient(p.Endpoints, tracer)
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
		ah.addAttr(&pb.Attribute{Id: option.name, Type: option.destType, Value: value})
	}
	return srcIpFound
}

func (p *PolicyPlugin) getAttrsFromEDNS0(ah *attrHolder, r *dns.Msg, ip string) {
	ipId := "source_ip"

	o := r.IsEdns0()
	if o == nil {
		ah.addAttr(&pb.Attribute{Id: ipId, Type: "address", Value: ip})
		return
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
	ah.addAttr(&pb.Attribute{Id: ipId, Type: "address", Value: ip})
	return
}

func (p *PolicyPlugin) retRcode(ah *attrHolder, w dns.ResponseWriter, r *dns.Msg, rcode int, err error) (int, error) {
	msg := &dns.Msg{}
	msg.SetRcode(r, rcode)
	return p.retAnswer(ah, w, msg, rcode, err)
}

func (p *PolicyPlugin) retAnswer(ah *attrHolder, w dns.ResponseWriter, r *dns.Msg, rcode int, err error) (int, error) {
	w.WriteMsg(r)
	if ah != nil && p.TapIO != nil && rcode != dns.RcodeRefused && rcode != dns.RcodeServerFailure {
		if pw := dnstap.NewProxyWriter(w); pw != nil {
			p.TapIO.SendCRExtraMsg(pw, ah.attributes())
		}
	}
	return rcode, err
}

func join(key, value string) string { return "," + key + ":" + value }

func (p *PolicyPlugin) retDebugInfo(ah *attrHolder, w dns.ResponseWriter,
	r *dns.Msg, resolve int) (int, error) {
	debugQueryInfo := resConv[resolve]

	if ah.resp1Beg > 0 {
		debugQueryInfo += join("query", ah.effect1.String())
		for _, item := range ah.resp1() {
			debugQueryInfo += join(item.Id, item.Value)
		}
	}
	if ah.resp2Beg > 0 {
		debugQueryInfo += join("response", ah.effect2.String())
		for _, item := range ah.resp2() {
			debugQueryInfo += join(item.Id, item.Value)
		}
	}

	hdr := dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassCHAOS, Ttl: 0}
	r.Answer = append(r.Answer, &dns.TXT{Hdr: hdr, Txt: []string{debugQueryInfo}})
	r.Response = true
	r.Rcode = dns.RcodeSuccess
	w.WriteMsg(r)
	return dns.RcodeSuccess, nil
}

func (p *PolicyPlugin) decodeDebugMsg(r *dns.Msg) (*dns.Msg, bool) {
	if r != nil && len(r.Question) > 0 {
		q := r.Question[0]
		if q.Qclass == dns.ClassCHAOS && q.Qtype == dns.TypeTXT {
			name := strings.ToLower(q.Name)
			if strings.HasSuffix(name, p.DebugSuffix) {
				req := new(dns.Msg)
				name = strings.TrimSuffix(name, p.DebugSuffix)
				req.SetQuestion(name, dns.TypeA)
				return req, true
			}
		}
	}
	return r, false
}

// ServeDNS implements the Handler interface.
func (p *PolicyPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	q, debugQuery := p.decodeDebugMsg(r)
	state := request.Request{W: w, Req: q}

	ah := newAttrHolder(state.Name(), state.QType())
	p.getAttrsFromEDNS0(ah, r, state.IP())

	// validate domain name (validation #1)
	if err := p.validate(ah); err != nil {
		return p.retRcode(nil, w, r, dns.RcodeServerFailure, err)
	}

	var (
		status  int
		respMsg *dns.Msg
		resolve int
		err     error
	)

	if ah.action == typeAllow {
		// resolve domain name to IP
		responseWriter := new(Writer)
		status, err = plugin.NextOrFailure(p.Name(), p.Next, ctx, responseWriter, state.Req)
		respMsg = responseWriter.Msg
		if err == nil {
			address := extractRespIP(respMsg)
			// if external resolver ret code is not RcodeSuccess
			// address is not filled from the answer
			// in this case just pass through answer w/o validation
			if len(address) > 0 {
				resolve = resolveYes
				ah.addAddress(address)
				// validate response IP (validation #2)
				err = p.validate(ah)
				if err != nil {
					return p.retRcode(nil, w, r, dns.RcodeServerFailure, err)
				}
			} else {
				resolve = resolveNo
			}
		} else {
			resolve = resolveFailed
		}
	} else {
		resolve = resolveSkip
	}

	if debugQuery && (ah.action != typeRefuse) {
		return p.retDebugInfo(ah, w, r, resolve)
	}

	if err != nil {
		return p.retRcode(nil, w, r, dns.RcodeServerFailure, err)
	}

	switch ah.action {
	case typeAllow:
		return p.retAnswer(ah, w, respMsg, status, nil)
	case typeRedirect:
		return p.redirect(ah, ctx, w, r, ah.redirect.Value)
	case typeBlock:
		return p.retRcode(ah, w, r, dns.RcodeNameError, nil)
	case typeRefuse:
		return p.retRcode(ah, w, r, dns.RcodeRefused, nil)
	}

	return p.retRcode(nil, w, r, dns.RcodeServerFailure, errInvalidAction)
}

// Name implements the Handler interface
func (p *PolicyPlugin) Name() string { return "policy" }

func (p *PolicyPlugin) redirect(ah *attrHolder, ctx context.Context, w dns.ResponseWriter, r *dns.Msg, dst string) (int, error) {
	state := request.Request{W: w, Req: r}
	var rr dns.RR
	cname := false

	if ipv4 := net.ParseIP(dst).To4(); ipv4 != nil {
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass()}
		rr.(*dns.A).A = ipv4
	} else if ipv6 := net.ParseIP(dst).To16(); ipv6 != nil {
		rr = new(dns.AAAA)
		rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeAAAA, Class: state.QClass()}
		rr.(*dns.AAAA).AAAA = ipv6
	} else {
		dst = strings.TrimSuffix(dst, ".") + "."
		rr = new(dns.CNAME)
		rr.(*dns.CNAME).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeCNAME, Class: state.QClass()}
		rr.(*dns.CNAME).Target = dst
		cname = true
	}

	msg := new(dns.Msg)
	msg.Compress = true
	msg.Authoritative = true
	msg.Answer = []dns.RR{rr}

	if cname {
		req := new(dns.Msg)
		req.SetQuestion(dst, state.QType())

		responseWriter := new(Writer)
		status, err := plugin.NextOrFailure(p.Name(), p.Next, ctx, responseWriter, req)
		if err != nil {
			return p.retRcode(nil, w, r, dns.RcodeServerFailure, err)
		}

		msg.Answer = append(msg.Answer, responseWriter.Msg.Answer...)
		msg.SetRcode(r, status)
	} else {
		msg.SetRcode(r, dns.RcodeSuccess)
	}

	state.SizeAndDo(msg)
	return p.retAnswer(ah, w, msg, msg.Rcode, nil)
}

func (p *PolicyPlugin) validate(ah *attrHolder) error {
	response := new(pb.Response)
	err := p.pdp.ModalValidate(pb.Request{Attributes: ah.request()}, response)
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s", err)
		return err
	}

	ah.addResponse(response)
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
	localAddr  net.Addr
	remoteAddr net.Addr
	Msg        *dns.Msg
}

func (w *Writer) Close() error                  { return nil }
func (w *Writer) TsigStatus() error             { return nil }
func (w *Writer) TsigTimersOnly(b bool)         { return }
func (w *Writer) Hijack()                       { return }
func (w *Writer) LocalAddr() net.Addr           { return w.localAddr }
func (w *Writer) RemoteAddr() net.Addr          { return w.remoteAddr }
func (w *Writer) WriteMsg(m *dns.Msg) error     { w.Msg = m; return nil }
func (w *Writer) Write(buf []byte) (int, error) { return len(buf), nil }
