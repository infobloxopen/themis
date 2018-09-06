package policy

import (
	"errors"
	"net"
	"testing"

	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

func TestGetNameAndType(t *testing.T) {
	fqdn := dns.Fqdn("example.com")
	m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)

	qName, qType := getNameAndType(m)
	if qName != fqdn {
		t.Errorf("expected %q as query name but got %q", fqdn, qName)
	}

	if qType != dns.TypeA {
		t.Errorf("expected %d as query type but got %d", dns.TypeA, qType)
	}

	fqdn = "."
	qName, qType = getNameAndType(nil)
	if qName != fqdn {
		t.Errorf("expected %q as query name but got %q", fqdn, qName)
	}

	if qType != dns.TypeNone {
		t.Errorf("expected %d as query type but got %d", dns.TypeNone, qType)
	}
}

func TestGetNameAndClass(t *testing.T) {
	fqdn := dns.Fqdn("example.com")
	m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)

	qName, qClass := getNameAndClass(m)
	if qName != fqdn {
		t.Errorf("expected %q as query name but got %q", fqdn, qName)
	}

	if qClass != dns.ClassINET {
		t.Errorf("expected %d as query class but got %d", dns.ClassINET, qClass)
	}

	fqdn = "."
	qName, qClass = getNameAndClass(nil)
	if qName != fqdn {
		t.Errorf("expected %q as query name but got %q", fqdn, qName)
	}

	if qClass != dns.ClassNONE {
		t.Errorf("expected %d as query class but got %d", dns.ClassNONE, qClass)
	}
}

func TestGetRemoteIP(t *testing.T) {
	w := testutil.NewTestAddressedNonwriter("192.0.2.1")
	a := getRemoteIP(w)
	if !a.Equal(net.ParseIP("192.0.2.1")) {
		t.Errorf("expected %s as remote address but got %s", "192.0.2.1", a)
	}

	w = testutil.NewTestAddressedNonwriter("192.0.2.1:53")
	a = getRemoteIP(w)
	if !a.Equal(net.ParseIP("192.0.2.1")) {
		t.Errorf("expected %s as remote address but got %s", "192.0.2.1", a)
	}
}

func TestGetRespIp(t *testing.T) {
	m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	testutil.AppendAnswer(m, testutil.NewA(net.ParseIP("192.0.2.1")))

	a := getRespIP(m)
	if !a.Equal(net.ParseIP("192.0.2.1")) {
		t.Errorf("expected %s as response address but got %s", "192.0.2.1", a)
	}

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeAAAA, dns.ClassINET)
	testutil.AppendAnswer(m, testutil.NewAAAA(net.ParseIP("2001:db8::1")))

	a = getRespIP(m)
	if !a.Equal(net.ParseIP("2001:db8::1")) {
		t.Errorf("expected %s as response address but got %s", "2001:db8::1", a)
	}

	m = testutil.MakeTestDNSMsg("www.example.com", dns.TypeCNAME, dns.ClassINET)
	testutil.AppendAnswer(m, testutil.NewCNAME("example.com"))

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

	m := testutil.MakeTestDNSMsgWithEdns0("example.com", dns.TypeA, dns.ClassINET,
		testutil.NewEdns0(
			testutil.NewEdns0Cookie("badc0de."),
			testutil.NewEdns0Local(0xfffd, []byte{0xde, 0xc0, 0xad, 0xb}),
			testutil.NewEdns0Local(0xfffe, []byte{0xef, 0xbe, 0xad, 0xde}),
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

func TestClearECS(t *testing.T) {
	m := testutil.MakeTestDNSMsgWithEdns0("example.com", dns.TypeA, dns.ClassINET,
		testutil.NewEdns0(
			testutil.NewEdns0Cookie("badc0de."),
			testutil.NewEdns0Subnet(net.ParseIP("192.0.2.1")),
			testutil.NewEdns0Local(0xfffe, []byte{0xb, 0xad, 0xc0, 0xde}),
			testutil.NewEdns0Subnet(net.ParseIP("2001:db8::1")),
		),
	)

	clearECS(m)
	testutil.AssertDNSMessage(t, "clearECS", 0, m, 0,
		";; opcode: QUERY, status: NOERROR, id: 0\n"+
			";; flags:; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 1\n\n"+
			";; QUESTION SECTION:\n"+
			";example.com.\tIN\t A\n\n"+
			";; ADDITIONAL SECTION:\n\n"+
			";; OPT PSEUDOSECTION:\n"+
			"; EDNS: version 0; flags: ; udp: 0\n"+
			"; COOKIE: badc0de.\n"+
			"; LOCAL OPT: 65534:0x0badc0de\n",
	)
}

func TestResetTTL(t *testing.T) {
	in := testutil.MakeTestDNSMsg("test.com", dns.TypeA, dns.ClassINET)

	out := new(dns.Msg)
	out.SetReply(in)

	rr, err := dns.NewRR("test.com.\t599\tIN\tA\t10.0.10.11\n")
	if err != nil {
		t.Errorf("Failed to create A record")
	}
	out.Answer = append(out.Answer, rr)

	rr, err = dns.NewRR("test.com.\t598\tIN\tAAAA\t10::12\n")
	if err != nil {
		t.Errorf("Failed to create AAAA record")
	}
	out.Answer = append(out.Answer, rr)

	rr, err = dns.NewRR("test.com.\t597\tIN\tNS\tns1.test.com\n")
	if err != nil {
		t.Errorf("Failed to create NS record")
	}
	out.Answer = append(out.Answer, rr)

	out = resetTTL(out)
	testutil.AssertDNSMessage(t, "resetTTL", 0, out, 0,
		";; opcode: QUERY, status: NOERROR, id: 0\n"+
			";; flags: qr; QUERY: 1, ANSWER: 3, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
			";; QUESTION SECTION:\n"+
			";test.com.\tIN\t A\n\n"+
			";; ANSWER SECTION:\n"+
			"test.com.\t0\tIN\tA\t10.0.10.11\n"+
			"test.com.\t0\tIN\tAAAA\t10::12\n"+
			"test.com.\t0\tIN\tNS\tns1.test.com.\n",
	)
}

func TestSetRedirectQueryAnswer(t *testing.T) {
	p := newPolicyPlugin()

	mp := &testutil.MockPlugin{
		Ip: net.ParseIP("192.0.2.153"),
		Rc: dns.RcodeSuccess,
	}
	p.next = mp

	m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w := testutil.NewTestAddressedNonwriter("192.0.2.1")

	rc, err := p.setRedirectQueryAnswer(context.TODO(), w, m, "192.0.2.53")
	if err != nil {
		t.Error(err)
	}
	testutil.AssertDNSMessage(t, "setRedirectQueryAnswer(192.0.2.53)", rc, m, dns.RcodeSuccess,
		";; opcode: QUERY, status: NOERROR, id: 0\n"+
			";; flags:; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
			";; QUESTION SECTION:\n"+
			";example.com.\tIN\t A\n\n"+
			";; ANSWER SECTION:\n"+
			"example.com.\t0\tIN\tA\t192.0.2.53\n",
	)

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	rc, err = p.setRedirectQueryAnswer(context.TODO(), w, m, "2001:db8::53")
	if err != nil {
		t.Error(err)
	}
	testutil.AssertDNSMessage(t, "setRedirectQueryAnswer(2001:db8::53)", rc, m, dns.RcodeSuccess,
		";; opcode: QUERY, status: NOERROR, id: 0\n"+
			";; flags:; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
			";; QUESTION SECTION:\n"+
			";example.com.\tIN\t A\n\n"+
			";; ANSWER SECTION:\n"+
			"example.com.\t0\tIN\tAAAA\t2001:db8::53\n",
	)

	m = testutil.MakeTestDNSMsg("redirect.example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	rc, err = p.setRedirectQueryAnswer(context.TODO(), w, m, "example.com")
	if err != nil {
		t.Error(err)
	}
	testutil.AssertDNSMessage(t, "setRedirectQueryAnswer(redirect.example.com->example.com)", rc, m, dns.RcodeSuccess,
		";; opcode: QUERY, status: NOERROR, id: 0\n"+
			";; flags: qr aa; QUERY: 1, ANSWER: 2, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
			";; QUESTION SECTION:\n"+
			";redirect.example.com.\tIN\t A\n\n"+
			";; ANSWER SECTION:\n"+
			"redirect.example.com.\t0\tIN\tCNAME\texample.com.\n"+
			"example.com.\t0\tIN\tA\t192.0.2.153\n",
	)

	m = new(dns.Msg)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	rc, err = p.setRedirectQueryAnswer(context.TODO(), w, m, "example.com")
	if err == nil {
		t.Errorf("expected errInvalidDNSMessage")
	} else if err != errInvalidDNSMessage {
		t.Errorf("expected errInvalidDNSMessage but got %T: %s", err, err)
	}

	testutil.AssertDNSMessage(t, "setRedirectQueryAnswer(empty)", rc, m, dns.RcodeServerFailure,
		";; opcode: QUERY, status: NOERROR, id: 0\n"+
			";; flags:; QUERY: 0, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n",
	)

	m = testutil.MakeTestDNSMsg("redirect.example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	errTest := errors.New("testError")
	mp.Err = errTest

	rc, err = p.setRedirectQueryAnswer(context.TODO(), w, m, "example.com")
	if err == nil {
		t.Errorf("expected errTest")
	} else if err != errTest {
		t.Errorf("expected errTest but got %T: %s", err, err)
	}

	testutil.AssertDNSMessage(t, "setRedirectQueryAnswer(redirect.example.com->error)", rc, m, dns.RcodeServerFailure,
		";; opcode: QUERY, status: NOERROR, id: 0\n"+
			";; flags:; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
			";; QUESTION SECTION:\n;"+
			"redirect.example.com.\tIN\t A\n",
	)
}
