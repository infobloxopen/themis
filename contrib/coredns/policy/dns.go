package policy

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/pkg/nonwriter"
	"github.com/miekg/dns"
)

var (
	errInvalidDNSMessage = errors.New("invalid DNS message")
	// Several rrCodes could be configured together separated by rrCodeDelimiter, each rrCode and it's record is separarated by rrCodePrefixDelimiter
	rrCodeDelimiter       = ";"
	rrCodePrefixDelimiter = "="
)

func getNameAndType(r *dns.Msg) (string, uint16) {
	if r == nil || len(r.Question) <= 0 {
		return ".", dns.TypeNone
	}

	q := r.Question[0]
	return q.Name, q.Qtype
}

func getNameAndClass(r *dns.Msg) (string, uint16) {
	if r == nil || len(r.Question) <= 0 {
		return ".", dns.ClassNONE
	}

	q := r.Question[0]
	return q.Name, q.Qclass
}

// helper function to get name, type and class from the question section of the request
func getNameTypeAndClass(r *dns.Msg) (string, uint16, uint16) {
	if r == nil || len(r.Question) <= 0 {
		return ".", dns.TypeNone, dns.ClassNONE
	}

	q := r.Question[0]
	return q.Name, q.Qtype, q.Qclass
}

func getRemoteIP(w dns.ResponseWriter) net.IP {
	addrPort := w.RemoteAddr().String()
	addr, _, err := net.SplitHostPort(w.RemoteAddr().String())
	if err != nil {
		addr = addrPort
	}

	return net.ParseIP(addr)
}

func getRespIP(r *dns.Msg) net.IP {
	if r == nil {
		return nil
	}

	var ip net.IP
	for _, rr := range r.Answer {
		switch rr := rr.(type) {
		case *dns.A:
			ip = rr.A
			break

		case *dns.AAAA:
			ip = rr.AAAA
			break
		}
	}

	return ip
}

func extractOptionsFromEDNS0(r *dns.Msg, optsMap map[uint16][]*edns0Opt, f func([]byte, []*edns0Opt)) {
	o := r.IsEdns0()
	if o == nil {
		return
	}

	var option []dns.EDNS0
	for _, o := range o.Option {
		if local, ok := o.(*dns.EDNS0_LOCAL); ok {
			if m, ok := optsMap[local.Code]; ok {
				f(local.Data, m)
				continue
			}
		}

		option = append(option, o)
	}

	o.Option = option
}

func clearECS(r *dns.Msg) {
	o := r.IsEdns0()
	if o == nil {
		return
	}

	option := make([]dns.EDNS0, 0, len(o.Option))
	for _, opt := range o.Option {
		if _, ok := opt.(*dns.EDNS0_SUBNET); !ok {
			option = append(option, opt)
		}
	}

	o.Option = option
}

func resetTTL(r *dns.Msg) *dns.Msg {
	if r == nil {
		return nil
	}

	for _, rr := range r.Answer {
		rr.Header().Ttl = 0
	}
	for _, rr := range r.Ns {
		rr.Header().Ttl = 0
	}

	return r
}

func (p *policyPlugin) setRedirectQueryAnswer(ctx context.Context, w dns.ResponseWriter, r *dns.Msg, dst string) (int, error) {
	if r == nil || len(r.Question) <= 0 {
		return dns.RcodeServerFailure, errInvalidDNSMessage
	}

	var rr dns.RR
	qName, qType, qClass := getNameTypeAndClass(r)
	record := findRecord(dst, qType)
	switch qType {
	case dns.TypeA, dns.TypeAAAA:
		ip := net.ParseIP(record)
		if rr = ip2rr(ip, qName, qClass); rr != nil {
			break
		}
		if cName := findRecord(dst, dns.TypeCNAME); cName != "" {
			return p.resolveByCname(ctx, w, r, cName, qName, qClass)
		}
	case dns.TypeTXT:
		if len(record) > 0 {
			rr = &dns.TXT{
				Hdr: dns.RR_Header{
					Name:   qName,
					Rrtype: dns.TypeTXT,
					Class:  qClass,
				},
				Txt: []string{record},
			}
		}
	case dns.TypeCNAME:
		if len(record) > 0 {
			rr = &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   qName,
					Rrtype: dns.TypeCNAME,
					Class:  qClass,
				},
				Target: record,
			}
		}
	}

	if rr != nil {
		r.Answer = []dns.RR{rr}
	}
	r.Rcode = dns.RcodeSuccess
	return r.Rcode, nil
}

// ip2rr returns a dns.A if ip is v4 and dns.AAAA if ip is v6, nil otherwise
func ip2rr(ip net.IP, name string, class uint16) dns.RR {
	if ipv4 := ip.To4(); ipv4 != nil {
		return &dns.A{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeA,
				Class:  class,
			},
			A: ipv4,
		}
	} else if ipv6 := ip.To16(); ipv6 != nil {
		return &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeAAAA,
				Class:  class,
			},
			AAAA: ipv6,
		}
	}
	return nil
}

// find record will return record for typ if dst is in rrCodeformat and has record for typ, if it does not have will send empty.
// if dst is not in rrCodeformat, will send entire string
func findRecord(dst string, typ uint16) string {
	var record string
	for _, each := range strings.Split(dst, rrCodeDelimiter) {
		entry := strings.SplitN(each, rrCodePrefixDelimiter, 2)
		if len(entry) != 2 || entry[0] == "" {
			return dst
		}
		if entry[0] == dns.TypeToString[typ] {
			record = entry[1]
		}
	}
	return record
}

func (p *policyPlugin) resolveByCname(ctx context.Context, w dns.ResponseWriter, r *dns.Msg, dst, name string, class uint16) (int, error) {
	if r == nil || len(r.Question) <= 0 {
		return dns.RcodeServerFailure, errInvalidDNSMessage
	}

	dst = dns.Fqdn(dst)
	rr := &dns.CNAME{
		Hdr: dns.RR_Header{
			Name:   name,
			Rrtype: dns.TypeCNAME,
			Class:  class,
		},
		Target: dst,
	}

	origName := name
	r.Question[0].Name = dst

	nw := nonwriter.New(w)

	if _, err := plugin.NextOrFailure(p.Name(), p.next, ctx, nw, r); err != nil {
		r.Question[0].Name = origName
		return dns.RcodeServerFailure, err
	}

	nw.Msg.CopyTo(r)
	r.Question[0].Name = origName

	r.Answer = append([]dns.RR{rr}, r.Answer...)
	r.Authoritative = true
	return r.Rcode, nil
}
