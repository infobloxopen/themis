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
	errInvalidDNSMessage         = errors.New("invalid DNS message")
	errInvalidRedirectActionData = errors.New("Invalid redirect data for incoming query type in policy config")
)

// rrCodePrefixes holds all the rrCodes that we check for in the action_data of redirect policy
// Several rrCodes could be configured together separated by rrCodeDelimiter
var (
	rrCodePrefixes = map[uint16]string{
		dns.TypeA:    "A:",
		dns.TypeAAAA: "AAAA:",
		dns.TypeTXT:  "TXT:",
	}
	rrCodeDelimiter = ";"
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
	return q.Name, q.Qtype, q.Qtype
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
	qName, qType, qClass := getNameTypeAndClass(r)
	rr, err := getRRByType(dst, qName, qType, qClass)
	if err != nil {
		return dns.RcodeServerFailure, err
	}

	// if not get rr by qType, then try to get it by the data format of dst
	if rr == nil {
		ip := net.ParseIP(dst)
		rr = ip2rr(ip, qName, qClass)

		if rr == nil {
			dst = dns.Fqdn(dst)
			rr = &dns.CNAME{
				Hdr: dns.RR_Header{
					Name:   qName,
					Rrtype: dns.TypeCNAME,
					Class:  qClass,
				},
				Target: dst,
			}

			if r == nil || len(r.Question) <= 0 {
				return dns.RcodeServerFailure, errInvalidDNSMessage
			}

			origName := qName
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
	}

	r.Answer = []dns.RR{rr}
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

// check if rrCode: prefix is present in dst and return all the values for each type. No validation of the action data is done here
func getRRCodePrefix(dst string) (map[uint16]string, bool) {
	redirectForRRCodes := map[uint16]string{}
	for _, each := range strings.Split(dst, ";") {
		for typ, prefix := range rrCodePrefixes {
			if strings.HasPrefix(each, prefix) {
				redirectForRRCodes[typ] = strings.TrimPrefix(each, prefix)
			}
		}
	}

	present := false
	if len(redirectForRRCodes) > 0 {
		present = true
	}
	return redirectForRRCodes, present
}

// try to get RR based on the typ
func getRRByType(dst, name string, typ, class uint16) (dns.RR, error) {
	var rr dns.RR
	redirectForRRCodes, present := getRRCodePrefix(dst)

	// if rrcode: is not present in redirect action_data, then return to get rr by data format only.
	if !present {
		return nil, nil
	}

	dest := redirectForRRCodes[typ]
	// if rrcode: is present in redirect action_data, but not for incoming query type, then cannot determine by type or dataformat also, so err out
	if present && len(dest) == 0 {
		return nil, errInvalidRedirectActionData
	}

	switch typ {
	case dns.TypeTXT:
		rr = &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   name,
				Rrtype: dns.TypeTXT,
				Class:  class,
			},
			Txt: []string{dest},
		}
	case dns.TypeA:
		if ip := net.ParseIP(dest); ip != nil && ip.To4() != nil {
			rr = &dns.A{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeA,
					Class:  class,
				},
				A: ip,
			}
		} else {
			return nil, errInvalidRedirectActionData
		}
	case dns.TypeAAAA:
		ip := net.ParseIP(dest)
		if ip != nil && ip.To16() != nil {
			rr = &dns.AAAA{
				Hdr: dns.RR_Header{
					Name:   name,
					Rrtype: dns.TypeAAAA,
					Class:  class,
				},
				AAAA: ip,
			}
		} else {
			return nil, errInvalidRedirectActionData
		}
	}

	return rr, nil
}
