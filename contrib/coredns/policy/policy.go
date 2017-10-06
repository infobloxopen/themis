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
	"time"

	"github.com/coredns/coredns/plugin"
	tapplg "github.com/coredns/coredns/plugin/dnstap"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/coredns/coredns/plugin/pkg/trace"
	"github.com/coredns/coredns/request"

	tap "github.com/infobloxopen/themis/contrib/coredns/policy/policytap"
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
	typeRefuse = iota
	typeAllow
	typeRedirect
	typeBlock
	typeInvalid
)

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
	options     map[uint16][]edns0Map
	TapIO       tapplg.IORoutine
	Trace       plugin.Handler
	Next        plugin.Handler
	pdp         pep.Client
	DebugSuffix string
	ErrorFunc   func(dns.ResponseWriter, *dns.Msg, int) // failover error handler
}

type Response struct {
	Action   int
	Redirect string
}

func makeResponse(resp *pb.Response) (ret Response) {
	if resp == nil {
		log.Printf("[ERROR] PDP response pointer is nil")
		ret.Action = typeInvalid
		return
	}
	if resp.Effect == pb.Response_PERMIT {
		ret.Action = typeAllow
		return
	} else if resp.Effect == pb.Response_DENY {
		for _, item := range resp.Obligation {
			switch item.Id {
			case "refuse":
				if item.Value == "true" {
					ret.Action = typeRefuse
					return
				}
			case "redirect_to":
				if item.Value != "" {
					ret.Action = typeRedirect
					ret.Redirect = item.Value
					return
				}
			}
		}
		ret.Action = typeBlock
		return
	} else {
		log.Printf("[ERROR] PDP Effect: %s", resp.Effect)
		ret.Action = typeInvalid
		return
	}
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
	p.options[ecode] = append(p.options[ecode], edns0Map{name, ednsType, destType, uint(size), uint(start), uint(end)})
	return nil
}

func parseHex(data []byte, option edns0Map) string {
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

func parseOptionGroup(data []byte, options []edns0Map) ([]*pb.Attribute, bool) {
	srcIpFound := false
	var attrs []*pb.Attribute
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
		attrs = append(attrs, &pb.Attribute{Id: option.name, Type: option.destType, Value: value})
	}
	return attrs, srcIpFound
}

func (p *PolicyPlugin) getAttrsFromEDNS0(r *dns.Msg, ip string) []*pb.Attribute {
	ipId := "source_ip"
	var attrs []*pb.Attribute

	o := r.IsEdns0()
	if o == nil {
		return []*pb.Attribute{{Id: ipId, Type: "address", Value: ip}}
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
		group, srcIpFound := parseOptionGroup(optLocal.Data, options)
		attrs = append(attrs, group...)
		if srcIpFound {
			ipId = "proxy_source_ip"
		}
	}
	attrs = append(attrs, &pb.Attribute{Id: ipId, Type: "address", Value: ip})
	return attrs
}

func (p *PolicyPlugin) retRcode(w dns.ResponseWriter, r *dns.Msg, rcode int, err error) (int, error) {
	msg := &dns.Msg{}
	msg.SetRcode(r, rcode)
	w.WriteMsg(msg)
	return rcode, err
}

func join(key, value string) string { return "," + key + ":" + value }

func (p *PolicyPlugin) retDebugInfo(r *dns.Msg, w dns.ResponseWriter,
	respDomain *pb.Response, respIP *pb.Response, resolve string) (int, error) {
	hdr := dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassCHAOS, Ttl: 0}
	debugQueryInfo := "resolve:" + resolve
	if respDomain != nil {
		debugQueryInfo += join("query", respDomain.Effect.String())
		for _, item := range respDomain.Obligation {
			debugQueryInfo += join(item.Id, item.Value)
		}
	}
	if respIP != nil {
		debugQueryInfo += join("response", respIP.Effect.String())
		for _, item := range respIP.Obligation {
			debugQueryInfo += join(item.Id, item.Value)
		}
	}
	r.Answer = append(r.Answer, &dns.TXT{Hdr: hdr, Txt: []string{debugQueryInfo}})
	r.Response = true
	r.Rcode = dns.RcodeSuccess
	w.WriteMsg(r)
	return dns.RcodeSuccess, nil
}

func (p *PolicyPlugin) handlePermit(ctx context.Context, w dns.ResponseWriter,
	r *dns.Msg, tapAttrs []*tap.DnstapAttribute, respDomain *pb.Response, debugQuery bool) (int, error) {
	req := r
	responseWriter := nonwriter.New(w)
	if debugQuery {
		req = new(dns.Msg)
		name := strings.TrimSuffix(r.Question[0].Name, p.DebugSuffix)
		req.SetQuestion(name, dns.TypeA)
	}
	status, err := plugin.NextOrFailure(p.Name(), p.Next, ctx, responseWriter, req)
	if err != nil {
		if debugQuery {
			return p.retDebugInfo(r, w, respDomain, nil, "failed")
		}
		return p.retRcode(w, r, dns.RcodeServerFailure, err)
	}

	var address string
	for _, rr := range responseWriter.Msg.Answer {
		switch rr.Header().Rrtype {
		case dns.TypeA, dns.TypeAAAA:
			if a, ok := rr.(*dns.A); ok {
				address = a.A.String()
			} else if a, ok := rr.(*dns.AAAA); ok {
				address = a.AAAA.String()
			}
		}
	}

	// if external resolver ret code is not RcodeSuccess
	// address is not filled from the answer
	// in this case just pass through answer w/o validation
	if address == "" {
		if debugQuery {
			return p.retDebugInfo(r, w, respDomain, nil, "no")
		}
		w.WriteMsg(responseWriter.Msg)
		return status, nil
	}

	attrs := []*pb.Attribute{{Id: "type", Type: "string", Value: "response"}}
	addressAttr := &pb.Attribute{Id: "address", Type: "address", Value: address}
	attrs = append(attrs, addressAttr)
	attrs = append(attrs, respDomain.Obligation...)

	response := new(pb.Response)
	err = p.pdp.Validate(ctx, pb.Request{Attributes: attrs}, response)
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s", err)
		return p.retRcode(w, r, dns.RcodeServerFailure, err)
	}

	if p.TapIO != nil {
		tapAttrs = append(tapAttrs, tap.PdpAttr2DnstapAttr(addressAttr))
		tap.SendPolicyHitMsg(p.TapIO, time.Now(), r, tap.PolicyHitMessage_POLICY_TRIGGER_ADDRESS, tapAttrs, response)
	}

	resp := makeResponse(response)

	if debugQuery && resp.Action != typeRefuse {
		return p.retDebugInfo(r, w, respDomain, response, "yes")
	}

	switch resp.Action {
	case typeAllow:
		w.WriteMsg(responseWriter.Msg)
		return status, nil
	case typeRedirect:
		return p.redirect(ctx, w, r, resp.Redirect)
	case typeBlock:
		return p.retRcode(w, r, dns.RcodeNameError, nil)
	case typeRefuse:
		return p.retRcode(w, r, dns.RcodeRefused, nil)
	}

	return p.retRcode(w, r, dns.RcodeServerFailure, errInvalidAction)
}

// ServeDNS implements the Handler interface.
func (p *PolicyPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	var (
		debugQuery bool
		domain     string
	)

	state := request.Request{W: w, Req: r}

	domain = state.Name()
	if state.QClass() == dns.ClassCHAOS && state.QType() == dns.TypeTXT {
		if strings.HasSuffix(domain, p.DebugSuffix) {
			debugQuery = true
			domain = strings.TrimSuffix(domain, p.DebugSuffix)
		}
	}

	// need to process OPT to get customer id
	attrs := []*pb.Attribute{
		{Id: "type", Type: "string", Value: "query"},
		{Id: "domain_name", Type: "domain", Value: strings.TrimRight(domain, ".")},
		{Id: "dns_qtype", Type: "string", Value: dns.TypeToString[state.QType()]},
	}

	ednsStart := len(attrs)
	ednsAttrs := p.getAttrsFromEDNS0(r, state.IP())
	attrs = append(attrs, ednsAttrs...)

	response := new(pb.Response)
	err := p.pdp.Validate(ctx, pb.Request{Attributes: attrs}, response)
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s", err)
		return p.retRcode(w, r, dns.RcodeServerFailure, err)
	}

	tapAttrs := []*tap.DnstapAttribute{}
	if p.TapIO != nil {
		tapAttrs = tap.ConvertAttrs(attrs[ednsStart:])
		tap.SendPolicyHitMsg(p.TapIO, time.Now(), r, tap.PolicyHitMessage_POLICY_TRIGGER_DOMAIN, tapAttrs, response)
	}

	resp := makeResponse(response)

	if debugQuery && (resp.Action == typeRedirect || resp.Action == typeBlock || resp.Action == typeInvalid) {
		return p.retDebugInfo(r, w, response, nil, "skip")
	}

	switch resp.Action {
	case typeAllow:
		return p.handlePermit(ctx, w, r, tapAttrs, response, debugQuery)
	case typeRedirect:
		return p.redirect(ctx, w, r, resp.Redirect)
	case typeBlock:
		return p.retRcode(w, r, dns.RcodeNameError, nil)
	case typeRefuse:
		return p.retRcode(w, r, dns.RcodeRefused, nil)
	}

	return p.retRcode(w, r, dns.RcodeServerFailure, errInvalidAction)
}

// Name implements the Handler interface
func (p *PolicyPlugin) Name() string { return "policy" }

func (p *PolicyPlugin) redirect(ctx context.Context, w dns.ResponseWriter, r *dns.Msg, dst string) (int, error) {
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

	a := new(dns.Msg)
	a.Compress = true
	a.Authoritative = true
	a.Answer = []dns.RR{rr}

	if cname {
		msg := new(dns.Msg)
		msg.SetQuestion(dst, state.QType())

		responseWriter := nonwriter.New(w)
		status, err := plugin.NextOrFailure(p.Name(), p.Next, ctx, responseWriter, msg)
		if err != nil {
			return p.retRcode(w, r, dns.RcodeServerFailure, err)
		}

		a.Answer = append(a.Answer, responseWriter.Msg.Answer...)
		a.SetRcode(r, status)
	} else {
		a.SetRcode(r, dns.RcodeSuccess)
	}

	state.SizeAndDo(a)
	w.WriteMsg(a)

	return a.Rcode, nil
}
