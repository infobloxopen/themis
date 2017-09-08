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
	Endpoints []string
	symbols   []edns0Map
	Trace     middleware.Handler
	Next      middleware.Handler
	pdp       pep.Client
	ErrorFunc func(dns.ResponseWriter, *dns.Msg, int) // failover error handler
}

// Response represents decision received from PDP server.
type Response struct {
	Permit   bool   `pdp:"Effect"`
	Redirect string `pdp:"redirect_to"`
	PolicyID string `pdp:"policy_id"`
	Refuse   string `pdp:"refuse"`
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

// NewLocalResponseWriter implements intermediate ResponseWriter to apply all
// other middleware effects before validating DNS reply.
type NewLocalResponseWriter struct {
	localAddr  net.Addr
	remoteAddr net.Addr
	Msg        *dns.Msg
}

// Write implements the dns.ResponseWriter interface.
func (r *NewLocalResponseWriter) Write(b []byte) (int, error) {
	r.Msg = new(dns.Msg)
	return len(b), r.Msg.Unpack(b)
}

// Close implements the dns.ResponseWriter interface.
func (r *NewLocalResponseWriter) Close() error { return nil }

// TsigStatus implements the dns.ResponseWriter interface.
func (r *NewLocalResponseWriter) TsigStatus() error { return nil }

// TsigTimersOnly implements the dns.ResponseWriter interface.
func (r *NewLocalResponseWriter) TsigTimersOnly(b bool) { return }

// Hijack implements the dns.ResponseWriter interface.
func (r *NewLocalResponseWriter) Hijack() { return }

// LocalAddr implements the dns.ResponseWriter interface.
func (r *NewLocalResponseWriter) LocalAddr() net.Addr { return r.localAddr }

// RemoteAddr implements the dns.ResponseWriter interface.
func (r *NewLocalResponseWriter) RemoteAddr() net.Addr { return r.remoteAddr }

// WriteMsg implements the dns.ResponseWriter interface.
func (r *NewLocalResponseWriter) WriteMsg(m *dns.Msg) error { r.Msg = m; return nil }

func (p *PolicyMiddleware) retRcode(w dns.ResponseWriter, r *dns.Msg, rcode int) (int, error) {
	msg := &dns.Msg{}
	msg.SetRcode(r, rcode)
	w.WriteMsg(msg)
	return rcode, nil
}

func (p *PolicyMiddleware) handlePermit(ctx context.Context, w dns.ResponseWriter, r *dns.Msg, attrs []*pb.Attribute) (int, error) {
	lw := new(NewLocalResponseWriter)
	status, err := middleware.NextOrFailure(p.Name(), p.Next, ctx, lw, r)
	if err != nil {
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
		w.WriteMsg(lw.Msg)
		return status, nil
	}

	attrs[0].Value = "response"
	attrs = append(attrs, &pb.Attribute{Id: "address", Type: "address", Value: address})

	var response Response
	err = p.pdp.Validate(ctx, pb.Request{Attributes: attrs}, &response)
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s\n", err)
		return dns.RcodeServerFailure, err
	}

	if response.Permit {
		w.WriteMsg(lw.Msg)
		return status, nil
	}

	if response.Redirect != "" {
		return p.redirect(ctx, w, r, response.Redirect)
	}

	return p.retRcode(w, r, dns.RcodeNameError)
}

// ServeDNS implements the Handler interface.
func (p *PolicyMiddleware) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	// need to process OPT to get customer id
	attrs := []*pb.Attribute{{Id: "type", Type: "string", Value: "query"}}

	if len(r.Question) > 0 {
		q := r.Question[0]
		attrs = append(attrs, &pb.Attribute{Id: "domain_name", Type: "domain", Value: strings.TrimRight(q.Name, ".")})
		attrs = append(attrs, &pb.Attribute{Id: "dns_qtype", Type: "string", Value: dns.TypeToString[q.Qtype]})
	}

	edns, foundSourceIP := p.getEDNS0Attrs(r)
	if len(edns) > 0 {
		attrs = append(attrs, edns...)
	}

	if foundSourceIP {
		attrs = append(attrs, &pb.Attribute{Id: "proxy_source_ip", Type: "address", Value: state.IP()})
	} else {
		attrs = append(attrs, &pb.Attribute{Id: "source_ip", Type: "address", Value: state.IP()})
	}

	var response Response
	err := p.pdp.Validate(ctx, pb.Request{Attributes: attrs}, &response)
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s\n", err)
		return dns.RcodeServerFailure, err
	}

	if response.Permit {
		attrs = append(attrs, &pb.Attribute{Id: "policy_id", Type: "string", Value: response.PolicyID})
		return p.handlePermit(ctx, w, r, attrs)
	}

	if response.Redirect != "" {
		return p.redirect(ctx, w, r, response.Redirect)
	}

	if response.Refuse == "true" {
		return p.retRcode(w, r, dns.RcodeRefused)
	}

	return p.retRcode(w, r, dns.RcodeNameError)
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

		lw := new(NewLocalResponseWriter)
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
