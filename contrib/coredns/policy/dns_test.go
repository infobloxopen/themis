package policy

import (
	"net"
	"testing"

	"github.com/miekg/dns"
)

func TestGetNameAndType(t *testing.T) {
	fqdn := dns.Fqdn("example.com")
	m := makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)

	qName, qType := getNameAndType(m)
	if qName != fqdn {
		t.Errorf("expected %q as query name but got %q", fqdn, qName)
	}

	if qType != dns.TypeA {
		t.Errorf("expected %d as query type but got %d", dns.TypeA, qType)
	}

	fqdn = dns.Fqdn("")
	qName, qType = getNameAndType(nil)
	if qName != fqdn {
		t.Errorf("expected %q as query name but got %q", fqdn, qName)
	}

	if qType != dns.TypeNone {
		t.Errorf("expected %d as query type but got %d", dns.TypeNone, qType)
	}
}

func TestGetRemoveIP(t *testing.T) {
	w := newTestAddressedNonwriter("192.0.2.1")
	a := getRemoteIP(w)
	if !a.Equal(net.ParseIP("192.0.2.1")) {
		t.Errorf("expected %s as remote address but got %s", "192.0.2.1", a)
	}

	w = newTestAddressedNonwriter("192.0.2.1:53")
	a = getRemoteIP(w)
	if !a.Equal(net.ParseIP("192.0.2.1")) {
		t.Errorf("expected %s as remote address but got %s", "192.0.2.1", a)
	}
}

func TestGetRespIp(t *testing.T) {
	m := makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	appendAnswer(m, newA(net.ParseIP("192.0.2.1")))

	a := getRespIP(m)
	if !a.Equal(net.ParseIP("192.0.2.1")) {
		t.Errorf("expected %s as response address but got %s", "192.0.2.1", a)
	}

	m = makeTestDNSMsg("example.com", dns.TypeAAAA, dns.ClassINET)
	appendAnswer(m, newAAAA(net.ParseIP("2001:db8::1")))

	a = getRespIP(m)
	if !a.Equal(net.ParseIP("2001:db8::1")) {
		t.Errorf("expected %s as response address but got %s", "2001:db8::1", a)
	}

	m = makeTestDNSMsg("www.example.com", dns.TypeCNAME, dns.ClassINET)
	appendAnswer(m, newCNAME("example.com"))

	a = getRespIP(m)
	if a != nil {
		t.Errorf("expected no response address but got %s", a)
	}

	a = getRespIP(nil)
	if a != nil {
		t.Errorf("expected no response address but got %s", a)
	}
}

func TestExtractOptionsFromEDNS0(t *testing.T) {
	optsMap := map[uint16][]*edns0Opt{
		0xfffe: {
			{
				name:     "test",
				dataType: typeEDNS0Bytes,
				size:     4,
			},
		},
	}

	m := makeTestDNSMsgWithEdns0("example.com", dns.TypeA, dns.ClassINET,
		newEdns0(
			newEdns0Cookie("badc0de."),
			newEdns0Local(0xfffd, []byte{0xde, 0xc0, 0xad, 0xb}),
			newEdns0Local(0xfffe, []byte{0xef, 0xbe, 0xad, 0xde}),
		),
	)

	n := 0
	extractOptionsFromEDNS0(m, optsMap, func(b []byte, opts []*edns0Opt) {
		n++

		if string(b) != string([]byte{0xef, 0xbe, 0xad, 0xde}) {
			t.Errorf("expected [% x] as EDNS0 data for option %d but got [% x]", []byte{0xef, 0xbe, 0xad, 0xde}, n, b)
		}

		if len(opts) != 1 || opts[0].name != "test" {
			t.Errorf("expected %q ENDS0 for option %d but got %+v", "test", n, opts)
		}
	})

	if n != 1 {
		t.Errorf("expected exactly one EDNS0 option but got %d", n)
	}

	o := m.IsEdns0()
	if o == nil {
		t.Error("expected ENDS0 options in DNS message")
	} else if len(o.Option) != 2 {
		t.Errorf("expected exactly %d options remaining but got %d", 2, len(o.Option))
	}
}

func makeTestDNSMsg(n string, t uint16, c uint16) *dns.Msg {
	out := new(dns.Msg)
	out.Question = make([]dns.Question, 1)
	out.Question[0] = dns.Question{
		Name:   dns.Fqdn(n),
		Qtype:  t,
		Qclass: c,
	}
	return out
}

func appendAnswer(m *dns.Msg, rr ...dns.RR) {
	if m.Answer == nil {
		m.Answer = []dns.RR{}
	}

	m.Answer = append(m.Answer, rr...)
}

func newA(a net.IP) dns.RR {
	out := new(dns.A)
	out.Hdr.Name = "."
	out.Hdr.Rrtype = dns.TypeA
	out.A = a

	return out
}

func newAAAA(a net.IP) dns.RR {
	out := new(dns.AAAA)
	out.Hdr.Name = "."
	out.Hdr.Rrtype = dns.TypeAAAA
	out.AAAA = a

	return out
}

func newCNAME(s string) dns.RR {
	out := new(dns.CNAME)
	out.Hdr.Name = "."
	out.Hdr.Rrtype = dns.TypeCNAME
	out.Target = dns.Fqdn(s)

	return out
}

func makeTestDNSMsgWithEdns0(n string, t uint16, c uint16, o ...*dns.OPT) *dns.Msg {
	out := makeTestDNSMsg(n, t, c)

	extra := make([]dns.RR, len(o))
	for i, o := range o {
		extra[i] = o
	}

	out.Extra = extra
	return out
}

func newEdns0(o ...dns.EDNS0) *dns.OPT {
	out := new(dns.OPT)
	out.Hdr.Name = "."
	out.Hdr.Rrtype = dns.TypeOPT
	out.Option = o

	return out
}

func copyEdns0(in ...*dns.OPT) []*dns.OPT {
	out := make([]*dns.OPT, len(in))
	for i, o := range in {
		out[i] = new(dns.OPT)
		out[i].Hdr = o.Hdr
		out[i].Option = make([]dns.EDNS0, len(o.Option))
		copy(out[i].Option, o.Option)
	}

	return out
}

func newEdns0Cookie(s string) dns.EDNS0 {
	out := new(dns.EDNS0_COOKIE)
	out.Code = dns.EDNS0COOKIE
	out.Cookie = s

	return out
}

func newEdns0Local(c uint16, b []byte) dns.EDNS0 {
	out := new(dns.EDNS0_LOCAL)
	out.Code = c
	out.Data = b

	return out
}

type testAddressedNonwriter struct {
	dns.ResponseWriter
	ra  *testUDPAddr
	Msg *dns.Msg
}

type testUDPAddr struct {
	addr string
}

func newTestAddressedNonwriter(ra string) *testAddressedNonwriter {
	return &testAddressedNonwriter{
		ResponseWriter: nil,
		ra:             newUDPAddr(ra),
	}
}

func (w *testAddressedNonwriter) RemoteAddr() net.Addr {
	return w.ra
}

func (w *testAddressedNonwriter) WriteMsg(res *dns.Msg) error {
	w.Msg = res
	return nil
}

func newUDPAddr(addr string) *testUDPAddr {
	return &testUDPAddr{
		addr: addr,
	}
}

func (a *testUDPAddr) String() string {
	return a.addr
}

func (a *testUDPAddr) Network() string {
	return "udp"
}
