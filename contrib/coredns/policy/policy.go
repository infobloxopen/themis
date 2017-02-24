package policy

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/coredns/coredns/middleware"
	"github.com/coredns/coredns/request"

	pb "github.com/infobloxopen/policy-box/pdp-service"
	"github.com/infobloxopen/policy-box/pep"
	"github.com/miekg/dns"
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
	Endpoint  string
	Zones     []string
	EDNS0Map  []edns0Map
	Timeout   time.Duration
	Next      middleware.Handler
	pdp       *pep.Client
	ErrorFunc func(dns.ResponseWriter, *dns.Msg, int) // failover error handler
}

func (p *PolicyMiddleware) Connect() error {
	log.Printf("[DEBUG] Connecting %v", p)
	p.pdp = pep.NewClient(p.Endpoint)
	return p.pdp.Connect(p.Timeout)
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

func (p *PolicyMiddleware) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	// need to process OPT to get customer id
	var attrs []*pb.Attribute
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

	var result pb.Response
	err := p.pdp.Validate(pb.Request{Attributes: attrs}, &result)
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s\n", err)
		return dns.RcodeServerFailure, err
	}

	rcode := dns.RcodeRefused
	switch result.Effect {
	case pb.Response_PERMIT:
		return middleware.NextOrFailure(p.Name(), p.Next, ctx, w, r)
	case pb.Response_DENY:
		if len(result.Obligation) > 0 {
			o := result.Obligation[0]
			if o.Id == "redirect_to" {
				return p.redirect(o.Value, w, r)
			} else {
				log.Printf("[WARNING] Unknown obligation: %v", o)
			}
		}
		if len(result.Obligation) > 1 {
			log.Printf("[WARNING] Only the first obligation will be enforced: %v", result.Obligation)
		}
	}

	return rcode, err
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
