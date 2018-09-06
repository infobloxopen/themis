package testutil

import (
	"context"
	"net"
	"sync/atomic"

	"github.com/miekg/dns"
)

const (
	MpModeConst = iota
	MpModeInc
	MpModeHalfInc
)

type MockPlugin struct {
	Ip   net.IP
	Err  error
	Rc   int
	Mode int
	Cnt  *uint32
}

// Name implements the plugin.Handler interface.
func (p *MockPlugin) Name() string {
	return "MockPlugin"
}

// ServeDNS implements the plugin.Handler interface.
func (p *MockPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	if p.Err != nil {
		return dns.RcodeServerFailure, p.Err
	}

	if r == nil || len(r.Question) <= 0 {
		return dns.RcodeServerFailure, nil
	}

	ip := p.Ip
	if p.Mode != MpModeConst && p.Cnt != nil {
		i := atomic.AddUint32(p.Cnt, 1)

		if p.Mode != MpModeHalfInc || i&1 == 0 {
			ip = addToIP(ip, i)
		}
	}

	q := r.Question[0]
	hdr := dns.RR_Header{
		Name:   q.Name,
		Rrtype: q.Qtype,
		Class:  q.Qclass,
	}

	if ipv4 := ip.To4(); ipv4 != nil {
		if q.Qtype != dns.TypeA {
			return dns.RcodeSuccess, nil
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Rcode = p.Rc

		if m.Rcode == dns.RcodeSuccess {
			m.Answer = append(m.Answer,
				&dns.A{
					Hdr: hdr,
					A:   ipv4,
				},
			)
		}

		w.WriteMsg(m)
	} else if ipv6 := ip.To16(); ipv6 != nil {
		if q.Qtype != dns.TypeAAAA {
			return dns.RcodeSuccess, nil
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true
		m.Rcode = p.Rc

		if m.Rcode == dns.RcodeSuccess {
			m.Answer = append(m.Answer,
				&dns.AAAA{
					Hdr:  hdr,
					AAAA: ipv6,
				},
			)
		}

		w.WriteMsg(m)
	}

	return p.Rc, nil
}

func addToIP(ip net.IP, n uint32) net.IP {
	if n == 0 {
		return ip
	}

	out := net.IP(make([]byte, len(ip)))
	copy(out, ip)

	d := uint(n % 256)
	n /= 256

	c := uint(n % 256)
	n /= 256

	b := uint(n % 256)
	n /= 256

	a := uint(n)

	t := uint(out[len(out)-1])
	t += d
	out[len(out)-1] = byte(t % 256)

	z := uint(out[len(out)-2])
	if t > 255 {
		z++
	}

	z += c
	out[len(out)-2] = byte(z % 256)

	y := uint(out[len(out)-3])
	if z > 255 {
		y++
	}

	y += b
	out[len(out)-3] = byte(y % 256)

	x := uint(out[len(out)-4])
	if y > 255 {
		x++
	}

	x += a
	out[len(out)-4] = byte(x % 256)

	return out
}
