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

	pb "github.com/infobloxopen/policy-box/pdp-service"
	"github.com/infobloxopen/policy-box/pep"

	"github.com/miekg/dns"

	ot "github.com/opentracing/opentracing-go"

	"golang.org/x/net/context"
)

const (
	EDNS0_MAP_DATA_TYPE_BYTES = iota
	EDNS0_MAP_DATA_TYPE_HEX   = iota
	EDNS0_MAP_DATA_TYPE_IP    = iota
)

var stringToEDNS0MapType = map[string]uint16{
	"bytes":   EDNS0_MAP_DATA_TYPE_BYTES,
	"hex":     EDNS0_MAP_DATA_TYPE_HEX,
	"address": EDNS0_MAP_DATA_TYPE_IP,
}

type edns0Map struct {
	code     uint16
	name     string
	dataType uint16
	destType string
}

type PolicyMiddleware struct {
	Endpoints []string
	Zones     []string
	EDNS0Map  []edns0Map
	Trace     middleware.Handler
	Next      middleware.Handler
	pdp       *pep.Client
	ErrorFunc func(dns.ResponseWriter, *dns.Msg, int) // failover error handler
}

type  Response struct {
	Permit   bool   `pdp:"Effect"`
	Redirect net.IP `pdp:"redirect_to"`
}

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

func (p *PolicyMiddleware) AddEDNS0Map(code, name, dataType, destType string) error {
	c, err := strconv.ParseUint(code, 0, 16)
	if err != nil {
		return fmt.Errorf("Could not parse EDNS0 code: %s", err)
	}
	ednsType, ok := stringToEDNS0MapType[dataType]
	if !ok {
		return fmt.Errorf("Invalid dataType for EDNS0 map: %s", dataType)
	}
	p.EDNS0Map = append(p.EDNS0Map, edns0Map{uint16(c), name, ednsType, destType})
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
			for _, m := range p.EDNS0Map {
				if m.code == e.Code {
					var value string
					switch m.dataType {
					case EDNS0_MAP_DATA_TYPE_BYTES:
						value = string(e.Data)
					case EDNS0_MAP_DATA_TYPE_HEX:
						value = hex.EncodeToString(e.Data)
					case EDNS0_MAP_DATA_TYPE_IP:
						ip := net.IP(e.Data)
						value = ip.String()
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


type NewLocalResponseWriter struct {
	localAddr  net.Addr
	remoteAddr net.Addr
	Msg        *dns.Msg
}


func (r *NewLocalResponseWriter) Write(b []byte) (int, error) {
	r.Msg = new(dns.Msg)
	return len(b), r.Msg.Unpack(b)
}

// These methods implement the dns.ResponseWriter interface from Go DNS.
func (r *NewLocalResponseWriter) Close() error              { return nil }
func (r *NewLocalResponseWriter) TsigStatus() error         { return nil }
func (r *NewLocalResponseWriter) TsigTimersOnly(b bool)     { return }
func (r *NewLocalResponseWriter) Hijack()                   { return }
func (r *NewLocalResponseWriter) LocalAddr() net.Addr       { return r.localAddr }
func (r *NewLocalResponseWriter) RemoteAddr() net.Addr      { return r.remoteAddr }
func (r *NewLocalResponseWriter) WriteMsg(m *dns.Msg) error { r.Msg = m; return nil }


func (p *PolicyMiddleware) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	// need to process OPT to get customer id
	attrs := []*pb.Attribute{&pb.Attribute{Id: "type", Type: "string", Value: "query"}}

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

	if response.Permit == true {
		lw := new(NewLocalResponseWriter)
		status,err := middleware.NextOrFailure(p.Name(), p.Next, ctx, lw, r)
		if err != nil {
			return status, err
		}

		attrs[0].Value = "response"
		for _, rr := range lw.Msg.Answer {
			switch rr.Header().Rrtype {
			case dns.TypeA, dns.TypeAAAA:
				if a, ok := rr.(*dns.A); ok {
					attrs = append(attrs, &pb.Attribute{Id: "address", Type: "address", Value: a.A.String()})
				} else if a, ok := rr.(*dns.AAAA); ok {
					attrs = append(attrs, &pb.Attribute{Id: "address", Type: "address", Value: a.AAAA.String()})
				}
			default:
			}
		}
		attrs = append(attrs, &pb.Attribute{Id: "policy_id", Type: "string", Value: ""})

		var lresponse Response

		err = p.pdp.Validate(ctx, pb.Request{Attributes: attrs}, &lresponse)

		if err == nil {
			if lresponse.Permit == true && lresponse.Redirect == nil {
				w.WriteMsg(lw.Msg)
				return status, nil
			}

			if lresponse.Redirect != nil {
				log.Printf("[INFO] Redirecting request")
				return p.redirect(lresponse.Redirect.String(), lw, lw.Msg)
			}

			if lresponse.Permit == false {
				return dns.RcodeRefused, fmt.Errorf("[ERROR] Request not permitted")
			}
		}
		return dns.RcodeRefused, nil
	}

	if response.Redirect != nil {
		return p.redirect(response.Redirect.String(), w, r)
	}

	return dns.RcodeRefused, nil
}

// Name implements the Handler interface
func (p *PolicyMiddleware) Name() string { return "policy" }

func (p *PolicyMiddleware) redirect(redirect_to string, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	a := new(dns.Msg)
	a.SetReply(r)
	a.Compress = true
	a.Authoritative = true

	var rr dns.RR

	switch state.Family() {
	case 1:
		rr = new(dns.A)
		rr.(*dns.A).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: state.QClass()}
		rr.(*dns.A).A = net.ParseIP(redirect_to).To4()
	case 2:
		rr = new(dns.AAAA)
		rr.(*dns.AAAA).Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeAAAA, Class: state.QClass()}
		rr.(*dns.AAAA).AAAA = net.ParseIP(redirect_to)
	}

	a.Answer = []dns.RR{rr}

	state.SizeAndDo(a)
	w.WriteMsg(a)

	return 0, nil
}
