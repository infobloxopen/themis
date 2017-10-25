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
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/coredns/coredns/request"

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

	actCount
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
	Endpoints     []string
	BatchInterval uint
	BatchLimit    uint
	options       map[uint16][]edns0Map
	TapIO         dnstap.DnstapSender
	Trace         plugin.Handler
	Next          plugin.Handler
	pdp           pep.Client
	DebugSuffix   string
	ErrorFunc     func(dns.ResponseWriter, *dns.Msg, int) // failover error handler
}

// Connect establishes connection to PDP server.
func (p *PolicyPlugin) Connect() error {
	log.Printf("[DEBUG] Endpoints: %v", p)
	p.pdp = pep.NewBalancedClient(p.Endpoints, p.BatchInterval, p.BatchLimit)
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

func parseOptionGroup(data []byte, options []edns0Map) ([]*pdp.Attribute, bool) {
	srcIpFound := false
	var attrs []*pdp.Attribute
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
		attrs = append(attrs, &pdp.Attribute{Id: option.name, Type: option.destType, Value: value})
	}
	return attrs, srcIpFound
}

func (p *PolicyPlugin) getAttrsFromEDNS0(r *dns.Msg, ip string) []*pdp.Attribute {
	ipId := "source_ip"
	var attrs []*pdp.Attribute

	o := r.IsEdns0()
	if o == nil {
		return []*pdp.Attribute{{Id: ipId, Type: "address", Value: ip}}
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
	attrs = append(attrs, &pdp.Attribute{Id: ipId, Type: "address", Value: ip})
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
	ah *attrHolder, err error, respMsg *dns.Msg, address string) (int, error) {
	resolve := "no"
	if err != nil {
		resolve = "failed"
	} else if respMsg == nil {
		resolve = "skip"
	} else if len(address) > 0 {
		resolve = "yes"
	}
	debugQueryInfo := "resolve:" + resolve

	if ah.resp1Beg > 0 {
		debugQueryInfo += join("query", pdp.EffectName(ah.effect1))
		for _, item := range ah.resp1() {
			debugQueryInfo += join(item.Id, item.Value)
		}
	}
	if ah.resp2Beg > 0 {
		debugQueryInfo += join("response", pdp.EffectName(ah.effect2))
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

	ah := newAttrHolder(strings.TrimRight(state.Name(), "."), fmt.Sprintf("%x", state.QType()))
	ah.addAttrs(p.getAttrsFromEDNS0(r, state.IP()))

	if p.TapIO != nil {
		if debugQuery == false {
			if pw := dnstap.NewProxyWriter(w); pw != nil {
				w = pw
				defer func() {
					switch ah.action {
					// TODO: implement "log" (or "inform") action and send the msg for it
					case typeBlock, typeRedirect, typeRefuse:
						attrs := ah.attributes()
						p.TapIO.SendCRExtraMsg(time.Now(), pw, attrs)
					}
				}()
			}
		}
	}

	// validate domain name (validation #1)
	err := p.validate(ctx, ah)
	if err != nil {
		return p.retRcode(w, r, dns.RcodeServerFailure, err)
	}

	var (
		address string
		respMsg *dns.Msg
		status  int
	)
	if ah.action == typeAllow {
		// resolve domain name to IP
		rw := nonwriter.New(w)
		status, err = plugin.NextOrFailure(p.Name(), p.Next, ctx, rw, state.Req)
		respMsg = rw.Msg

		if err == nil && respMsg != nil {
			address = extractRespIP(respMsg)
			// if external resolver ret code is not RcodeSuccess
			// address is not filled from the answer
			// in this case just pass through answer w/o validation
			if len(address) > 0 {
				ah.addAddress(address)
				// validate response IP (validation #2)
				err = p.validate(ctx, ah)
				if err != nil {
					return p.retRcode(w, r, dns.RcodeServerFailure, err)
				}
			}
		}
	}

	if debugQuery && ah.action != typeRefuse {
		return p.retDebugInfo(r, w, ah, err, respMsg, address)
	}
	if err != nil {
		return p.retRcode(w, r, dns.RcodeServerFailure, err)
	}

	switch ah.action {
	case typeAllow:
		w.WriteMsg(respMsg)
		return status, nil
	case typeRedirect:
		return p.redirect(ctx, w, r, ah.redirect.Value)
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

func (p *PolicyPlugin) validate(ctx context.Context, ah *attrHolder) error {
	response, err := p.pdp.Validate(&pdp.Request{Attributes: ah.request()})
	if err != nil {
		log.Printf("[ERROR] Policy validation failed due to error %s", err)
		return err
	}

	ah.addResponse(response)
	return nil
}

func extractRespIP(m *dns.Msg) string {
	var address string
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
	return address
}
