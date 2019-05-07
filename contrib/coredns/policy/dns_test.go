package policy

import (
	"context"
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	"github.com/miekg/dns"
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

func TestGetNameTypeAndClass(t *testing.T) {
	fqdn := dns.Fqdn("example.com")
	m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)

	qName, qType, qClass := getNameTypeAndClass(m)
	if qName != fqdn {
		t.Errorf("expected %q as query name but got %q", fqdn, qName)
	}

	if qType != dns.TypeA {
		t.Errorf("expected %d as query type but got %d", dns.TypeA, qType)
	}

	if qClass != dns.ClassINET {
		t.Errorf("expected %d as query class but got %d", dns.ClassINET, qClass)
	}

	fqdn = "."
	qName, qType, qClass = getNameTypeAndClass(nil)
	if qName != fqdn {
		t.Errorf("expected %q as query name but got %q", fqdn, qName)
	}

	if qType != dns.TypeNone {
		t.Errorf("expected %d as query type but got %d", dns.TypeA, qType)
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

func TestIp2rr(t *testing.T) {
	testCases := []struct {
		ip    net.IP
		name  string
		class uint16
		exp   string
	}{
		{net.ParseIP("192.0.0.44"), "amazon.com", dns.ClassINET, "amazon.com	0	IN	A	192.0.0.44"},
		{net.ParseIP("2001:db8::68"), "google.com", dns.ClassCHAOS, "google.com	0	CH	AAAA	2001:db8::68"},
		{net.ParseIP(""), "hotstar.com", dns.ClassANY, ""},
	}

	for _, tc := range testCases {
		act := ip2rr(tc.ip, tc.name, tc.class)
		if len(tc.exp) == 0 && act != nil {
			t.Errorf("Expected no ip, got %s", act)
		} else if len(tc.exp) > 0 && (act == nil || act.String() != tc.exp) {
			t.Errorf("Expected %s, got %s", tc.exp, act)
		}
	}
}

func TestGetRRCodesPrefix(t *testing.T) {
	testCases := []struct {
		dst string
		exp map[uint16]string
	}{
		{"A:1.1.1.1", map[uint16]string{dns.TypeA: "1.1.1.1"}},
		{"AAAA:23ef:3546::8732", map[uint16]string{dns.TypeAAAA: "23ef:3546::8732"}},
		{"TXT:dummytoken", map[uint16]string{dns.TypeTXT: "dummytoken"}},
		{"1.1.1.1", nil},
		{"", nil},
	}

	for idx, tc := range testCases {
		act := getRRCodePrefix(tc.dst)
		if !reflect.DeepEqual(act, tc.exp) {
			t.Errorf("expected %v but got %v : %d", tc.exp, act, idx)
		}
	}
}

func TestGetRRByType(t *testing.T) {
	testCases := []struct {
		dst,
		name string
		typ,
		class uint16
		err error
		exp dns.RR
	}{
		// check for TXT action_data with TXT query type
		{"TXT:dummytoken", "test.net", dns.TypeTXT, dns.ClassINET, nil, &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   "test.net",
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
			},
			Txt: []string{"dummytoken"},
		},
		},
		// check for A action_data with A query type
		{"A:1.1.1.1", "test.net", dns.TypeA, dns.ClassINET, nil, &dns.A{
			Hdr: dns.RR_Header{
				Name:   "test.net",
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
			},
			A: net.ParseIP("1.1.1.1"),
		},
		},
		// check for AAAA action_data with AAAA query type
		{"AAAA:23ef:3546::8732", "test.net", dns.TypeAAAA, dns.ClassINET, nil, &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   "test.net",
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
			},
			AAAA: net.ParseIP("23ef:3546::8732"),
		},
		},
		// check for all action_data with TXT query type
		{"A:1.1.1.1;AAAA:23ef:3546::8732;TXT:dummytoken", "test.net", dns.TypeTXT, dns.ClassINET, nil, &dns.TXT{
			Hdr: dns.RR_Header{
				Name:   "test.net",
				Rrtype: dns.TypeTXT,
				Class:  dns.ClassINET,
			},
			Txt: []string{"dummytoken"},
		},
		},
		// check for all action_data with A query type
		{"A:1.1.1.1;AAAA:23ef:3546::8732;TXT:dummytoken", "test.net", dns.TypeA, dns.ClassINET, nil, &dns.A{
			Hdr: dns.RR_Header{
				Name:   "test.net",
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
			},
			A: net.ParseIP("1.1.1.1"),
		},
		},
		// check for all action_data with AAAA query type
		{"A:1.1.1.1;AAAA:23ef:3546::8732;TXT:dummytoken", "test.net", dns.TypeAAAA, dns.ClassINET, nil, &dns.AAAA{
			Hdr: dns.RR_Header{
				Name:   "test.net",
				Rrtype: dns.TypeAAAA,
				Class:  dns.ClassINET,
			},
			AAAA: net.ParseIP("23ef:3546::8732"),
		},
		},
		// check for all action_data with diff query type(non-supported)
		{"A:1.1.1.1;AAAA:23ef:3546::8732;TXT:dummytoken", "test.net", dns.TypeAVC, dns.ClassINET, errInvalidRedirectActionData, nil},
		// check for no AAAA action_data with AAAA query type
		{"A:1.1.1.1;TXT:dummytoken", "test.net", dns.TypeAAAA, dns.ClassINET, errInvalidRedirectActionData, nil},
		// check for all old format action_data with A query type(for backward compatibility)
		{"1.1.1.1", "test.net", dns.TypeA, dns.ClassINET, nil, nil},
	}

	for idx, tc := range testCases {
		act, err := getRRByType(tc.dst, tc.name, tc.typ, tc.class)
		if err != tc.err {
			t.Errorf("expected %v but got %v : %d", tc.err, err, idx)
		}
		if !reflect.DeepEqual(act, tc.exp) {
			t.Errorf("expected %v but got %v : %d", tc.exp, act, idx)
		}
	}
}
