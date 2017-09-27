// Package policy implements Policy Enforcement Point (PEP) for CoreDNS.
// It allows to connect CoreDNS to Policy Decision Point (PDP), send decision
// requests to the PDP and perform actions on DNS requests and replies according
// recieved decisions.
package policy

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/coredns/coredns/plugin/pkg/trace"
	"github.com/coredns/coredns/request"

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
)

var stringToEDNS0MapType = map[string]uint16{
	"bytes":   typeEDNS0Bytes,
	"hex":     typeEDNS0Hex,
	"address": typeEDNS0IP,
}

type edns0Map struct {
	name         string
	dataType     uint16
	destType     string
	stringOffset int
	stringSize   int
}

// PolicyPlugin represents a plugin instance that can validate DNS
// requests and replies using PDP server.
type PolicyPlugin struct {
	Endpoints   []string
	options     map[uint16]edns0Map
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

func makeResponse(resp pb.Response) (ret Response) {
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
		log.Printf("[ERROR] PDP Effect %s", resp.Effect)
		ret.Action = typeRefuse
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
	stringOffset, stringSize string) error {
	c, err := strconv.ParseUint(code, 0, 16)
	if err != nil {
		return fmt.Errorf("Could not parse EDNS0 code: %s", err)
	}
	offset, err := strconv.Atoi(stringOffset)
	if err != nil {
		return fmt.Errorf("Could not parse EDNS0 string offset: %s", err)
	}
	size, err := strconv.Atoi(stringSize)
	if err != nil {
		return fmt.Errorf("Could not parse EDNS0 string size: %s", err)
	}
	ednsType, ok := stringToEDNS0MapType[dataType]
	if !ok {
		return fmt.Errorf("Invalid dataType for EDNS0 map: %s", dataType)
	}
	ecode := uint16(c)
	if _, ok := p.options[ecode]; ok {
		return fmt.Errorf("Duplicated EDNS0 code: %d", ecode)
	}
	p.options[ecode] = edns0Map{name, ednsType, destType, offset, size}
	return nil
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
		option, ok := p.options[optLocal.Code]
		if !ok {
			continue
		}
		var value string
		switch option.dataType {
		case typeEDNS0Bytes:
			value = string(optLocal.Data)
		case typeEDNS0Hex:
			value = hex.EncodeToString(optLocal.Data)
		case typeEDNS0IP:
			ip := net.IP(optLocal.Data)
			value = ip.String()
		}
		from := option.stringOffset
		to := option.stringOffset + option.stringSize
		if to > 0 && to <= len(value) && from < to {
			value = value[from:to]
		}
		if option.name == "source_ip" {
			ipId = "proxy_source_ip"
		}
		attrs = append(attrs, &pb.Attribute{Id: option.name, Type: option.destType, Value: value})
	}
	attrs = append(attrs, &pb.Attribute{Id: ipId, Type: "address", Value: ip})
	return attrs
}

func (p *PolicyPlugin) retRcode(w dns.ResponseWriter, r *dns.Msg, rcode int) (int, error) {
	msg := &dns.Msg{}
	msg.SetRcode(r, rcode)
	w.WriteMsg(msg)
	return rcode, nil
}

func (p *PolicyPlugin) retDebugInfo(r *dns.Msg, w dns.ResponseWriter,
	resp pb.Response, resolve string) (int, error) {
	hdr := dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassCHAOS, Ttl: 0}
	debugQueryInfo := fmt.Sprintf("resolve: %s effect: %s", resolve, resp.Effect)
	for _, item := range resp.Obligation {
		debugQueryInfo += fmt.Sprintf(" %s: %s", item.Id, item.Value)
	}
	r.Answer = append(r.Answer, &dns.TXT{Hdr: hdr, Txt: []string{debugQueryInfo}})
	r.Response = true
	r.Rcode = dns.RcodeSuccess
	w.WriteMsg(r)
	return dns.RcodeSuccess, nil
}

func (p *PolicyPlugin) handlePermit(ctx context.Context, w dns.ResponseWriter,
	r *dns.Msg, attrs []*pb.Attribute, respDomain pb.Response, debugQuery bool) (int, error) {
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
			return p.retDebugInfo(r, w, respDomain, "failed")
		}
		return status, err
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
			return p.retDebugInfo(r, w, respDomain, "no")
		}
		w.WriteMsg(responseWriter.Msg)
		return status, nil
	}

	attrs[0].Value = "response"
	attrs = append(attrs, &pb.Attribute{Id: "address", Type: "address", Value: address})
	attrs = append(attrs, respDomain.Obligation...)

	var response pb.Response
	err = p.pdp.Validate(ctx, pb.Request{Attributes: attrs}, &response)
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s", err)
		return dns.RcodeServerFailure, err
	}

	resp := makeResponse(response)

	if debugQuery && resp.Action != typeRefuse {
		return p.retDebugInfo(r, w, response, "yes")
	}

	switch resp.Action {
	case typeAllow:
		w.WriteMsg(responseWriter.Msg)
		return status, nil
	case typeRedirect:
		return p.redirect(ctx, w, r, resp.Redirect)
	case typeBlock:
		return p.retRcode(w, r, dns.RcodeNameError)
	}

	return p.retRcode(w, r, dns.RcodeRefused)
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

	ednsAttrs := p.getAttrsFromEDNS0(r, state.IP())
	attrs = append(attrs, ednsAttrs...)

	var response pb.Response
	err := p.pdp.Validate(ctx, pb.Request{Attributes: attrs}, &response)
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s", err)
		return dns.RcodeServerFailure, err
	}

	resp := makeResponse(response)

	if debugQuery && (resp.Action == typeRedirect || resp.Action == typeBlock) {
		return p.retDebugInfo(r, w, response, "skip")
	}

	switch resp.Action {
	case typeAllow:
		return p.handlePermit(ctx, w, r, attrs, response, debugQuery)
	case typeRedirect:
		return p.redirect(ctx, w, r, resp.Redirect)
	case typeBlock:
		return p.retRcode(w, r, dns.RcodeNameError)
	}

	return p.retRcode(w, r, dns.RcodeRefused)
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
			return status, err
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
