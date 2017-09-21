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

	"github.com/coredns/coredns/middleware"
	"github.com/coredns/coredns/middleware/pkg/nonwriter"
	"github.com/coredns/coredns/middleware/pkg/trace"
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
	code         uint16
	name         string
	dataType     uint16
	destType     string
	stringOffset int
	stringSize   int
}

// PolicyMiddleware represents a middleware instance that can validate DNS
// requests and replies using PDP server.
type PolicyMiddleware struct {
	Endpoints   []string
	symbols     []edns0Map
	Trace       middleware.Handler
	Next        middleware.Handler
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
func (p *PolicyMiddleware) Connect() error {
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
func (p *PolicyMiddleware) Close() {
	p.pdp.Close()
}

// AddEDNS0Map adds new EDNS0 to table.
func (p *PolicyMiddleware) AddEDNS0Map(code, name, dataType, destType,
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
	p.symbols = append(p.symbols, edns0Map{uint16(c), name, ednsType, destType, offset, size})
	return nil
}

func (p *PolicyMiddleware) getEDNS0Attrs(r *dns.Msg) ([]*pb.Attribute, bool) {
	foundSourceIP := false
	var attrs []*pb.Attribute

	o := r.IsEdns0()
	if o == nil {
		return nil, false
	}

	for _, s := range o.Option {
		switch e := s.(type) {
		case *dns.EDNS0_NSID:
			// do stuff with e.Nsid
		case *dns.EDNS0_SUBNET:
			// access e.Family, e.Address, etc.
		case *dns.EDNS0_LOCAL:
			for _, m := range p.symbols {
				if m.code == e.Code {
					var value string
					switch m.dataType {
					case typeEDNS0Bytes:
						value = string(e.Data)
					case typeEDNS0Hex:
						value = hex.EncodeToString(e.Data)
					case typeEDNS0IP:
						ip := net.IP(e.Data)
						value = ip.String()
					}
					from := m.stringOffset
					to := m.stringOffset + m.stringSize
					if to > 0 && to <= len(value) && from < to {
						value = value[from:to]
					}
					foundSourceIP = foundSourceIP || (m.name == "source_ip")
					attrs = append(attrs, &pb.Attribute{Id: m.name, Type: m.destType, Value: value})
					break
				}
			}
		}
	}
	return attrs, foundSourceIP
}

func (p *PolicyMiddleware) retRcode(w dns.ResponseWriter, r *dns.Msg, rcode int) (int, error) {
	msg := &dns.Msg{}
	msg.SetRcode(r, rcode)
	w.WriteMsg(msg)
	return rcode, nil
}

func (p *PolicyMiddleware) retDebugInfo(r *dns.Msg, w dns.ResponseWriter,
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
	log.Printf("[DEBUG] Query \"%s\" response [%s]", r.Question[0].Name, debugQueryInfo)
	return dns.RcodeSuccess, nil
}

func (p *PolicyMiddleware) handlePermit(ctx context.Context, w dns.ResponseWriter,
	r *dns.Msg, attrs []*pb.Attribute, respDomain pb.Response, debugQuery bool) (int, error) {
	req := r
	lw := nonwriter.New(w)
	if debugQuery {
		req = new(dns.Msg)
		name := strings.TrimSuffix(r.Question[0].Name, p.DebugSuffix)
		req.SetQuestion(name, dns.TypeA)
	}
	status, err := middleware.NextOrFailure(p.Name(), p.Next, ctx, lw, req)
	if err != nil {
		if debugQuery {
			return p.retDebugInfo(r, w, respDomain, "failed")
		}
		return status, err
	}

	var address string
	for _, rr := range lw.Msg.Answer {
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
		w.WriteMsg(lw.Msg)
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
		w.WriteMsg(lw.Msg)
		return status, nil
	case typeRedirect:
		return p.redirect(ctx, w, r, resp.Redirect)
	case typeBlock:
		return p.retRcode(w, r, dns.RcodeNameError)
	}

	return p.retRcode(w, r, dns.RcodeRefused)
}

// ServeDNS implements the Handler interface.
func (p *PolicyMiddleware) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
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

	edns, foundSourceIP := p.getEDNS0Attrs(r)
	attrs = append(attrs, edns...)

	if foundSourceIP {
		attrs = append(attrs, &pb.Attribute{Id: "proxy_source_ip", Type: "address", Value: state.IP()})
	} else {
		attrs = append(attrs, &pb.Attribute{Id: "source_ip", Type: "address", Value: state.IP()})
	}

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
func (p *PolicyMiddleware) Name() string { return "policy" }

func (p *PolicyMiddleware) redirect(ctx context.Context, w dns.ResponseWriter, r *dns.Msg, dst string) (int, error) {
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

		lw := nonwriter.New(w)
		status, err := middleware.NextOrFailure(p.Name(), p.Next, ctx, lw, msg)
		if err != nil {
			return status, err
		}

		a.Answer = append(a.Answer, lw.Msg.Answer...)
		a.SetRcode(r, status)
	} else {
		a.SetRcode(r, dns.RcodeSuccess)
	}

	state.SizeAndDo(a)
	w.WriteMsg(a)

	return a.Rcode, nil
}
