package policy

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/coredns/coredns/plugin/dnstap/taprw"
	dtest "github.com/coredns/coredns/plugin/dnstap/test"
	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	"github.com/infobloxopen/themis/pdp"
	"github.com/miekg/dns"
	"golang.org/x/net/context"

	pb "github.com/infobloxopen/themis/contrib/coredns/policy/dnstap"
)

func TestNewPolicyPlugin(t *testing.T) {
	p := newPolicyPlugin()
	if p == nil {
		t.Error("can't create new policy plugin instance")
	}
}

func TestPolicyPluginName(t *testing.T) {
	p := newPolicyPlugin()

	n := p.Name()
	if n != "policy" {
		t.Errorf("expected %q as plugin name but got %q", "policy", n)
	}
}

func TestPolicyPluginValidate(t *testing.T) {
	p := newPolicyPlugin()
	p.conf.attrs.parseAttrList(attrListTypeVal1, `type="query"`, attrNameDomainName, attrNameSourceIP)

	m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w := testutil.NewTestAddressedNonwriter("192.0.2.1")

	ah := newAttrHolder(nil, p.conf.attrs)
	ah.addDnsQuery(w, m, p.conf.options)

	mpc := testutil.NewMockPdpClient(t)
	mpc.In = []pdp.AttributeAssignment{
		pdp.MakeStringAssignment("type", "query"),
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
		pdp.MakeAddressAssignment(attrNameSourceIP, net.ParseIP("192.0.2.1")),
	}
	mpc.Out = []pdp.AttributeAssignment{
		pdp.MakeStringAssignment(attrNameRedirectTo, "192.0.2.54"),
		pdp.MakeIntegerAssignment("policy_action", 4),
	}
	mpc.Effect = pdp.EffectPermit
	p.pdp = mpc

	goon, err := p.validate(nil, ah, attrListTypeVal1, nil)
	if !goon {
		t.Errorf("Unexpected result of validate(): expected true, but got %t", goon)
	}
	if err != nil {
		t.Errorf("Unexpected error of validate(): expected nil, but got %s", err)
	}
	testutil.AssertAttrList(t, ah.attrs,
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeTestDomain(dns.Fqdn("example.com"))),
		pdp.MakeIntegerAssignment(attrNameDNSQtype, int64(dns.TypeA)),
		pdp.MakeAddressAssignment(attrNameSourceIP, net.ParseIP("192.0.2.1")),
		emptyAttr,
		pdp.MakeIntegerAssignment("policy_action", 4),
		pdp.MakeStringAssignment(attrNameRedirectTo, "192.0.2.54"),
		emptyAttr,
		pdp.MakeStringAssignment("type", "query"),
	)
}

func TestPolicyPluginServeDNS(t *testing.T) {
	endpoint := "127.0.0.1:5555"
	srv := testutil.StartPDPServer(t, serveDNSTestPolicy, endpoint)
	defer func() {
		if logs := srv.Stop(); len(logs) > 0 {
			t.Logf("server logs:\n%s", logs)
		}
	}()

	if err := testutil.WaitForPortOpened(endpoint); err != nil {
		t.Fatalf("can't connect to PDP server: %s", err)
	}

	p := newPolicyPlugin()
	p.conf.endpoints = []string{endpoint}
	p.conf.connTimeout = time.Second
	p.conf.streams = 1
	p.conf.log = true
	p.conf.debugID = "<DEBUG>"
	p.conf.debugSuffix = "debug.local."
	p.conf.autoResAttrs = true
	p.conf.attrs.parseAttrList(attrListTypeVal1, attrNameDomainName, `type="query"`)
	p.conf.attrs.parseAttrList(attrListTypeVal2, attrNameAddress, `type="response"`)

	mp := &testutil.MockPlugin{
		Ip: net.ParseIP("192.0.2.53"),
		Rc: dns.RcodeSuccess,
	}
	p.next = mp

	g := testutil.NewLogGrabber()
	if err := p.connect(); err != nil {
		logs := g.Release()
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		t.Fatal(err)
	}
	defer p.closeConn()

	m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w := testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err := p.ServeDNS(context.TODO(), w, m)
	logs := g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr aa; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n\n"+
				";; ANSWER SECTION:\n"+
				"example.com.\t0\tIN\tA\t192.0.2.53\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	m = testutil.MakeTestDNSMsg("example.com.debug.local", dns.TypeTXT, dns.ClassCHAOS)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(debug)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.debug.local.\tCH\t TXT\n\n"+
				";; ANSWER SECTION:\n"+
				"example.com.debug.local.\t0\tCH\tTXT\t\"Ident: <DEBUG>\" "+
				"\"PDP response {Effect: Permit, Obligations: [policy_action: allow]}\" "+
				"\"PDP response {Effect: Permit, Obligations: [policy_action: allow]}\" "+
				"\"Domain resolution: resolved\"\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	m = testutil.MakeTestDNSMsg("example.redirect", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(domain redirect)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.redirect.\tIN\t A\n\n"+
				";; ANSWER SECTION:\n"+
				"example.redirect.\t0\tIN\tA\t192.0.2.254\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	mp.Ip = net.ParseIP("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(address redirect)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n\n"+
				";; ANSWER SECTION:\n"+
				"example.com.\t0\tIN\tA\t192.0.2.253\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	mp.Ip = net.ParseIP("192.0.2.53")

	m = testutil.MakeTestDNSMsg("example.block", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(domain block)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NXDOMAIN, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.block.\tIN\t A\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	mp.Ip = net.ParseIP("192.0.2.17")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(address block)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NXDOMAIN, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	mp.Ip = net.ParseIP("192.0.2.53")

	m = testutil.MakeTestDNSMsg("example.refuse", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(domain refuse)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: REFUSED, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.refuse.\tIN\t A\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	mp.Ip = net.ParseIP("192.0.2.33")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(address refuse)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: REFUSED, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	mp.Ip = net.ParseIP("192.0.2.53")

	m = testutil.MakeTestDNSMsg("example.drop", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(domain drop)", rc, w.Msg, dns.RcodeSuccess,
			"<nil> MsgHdr",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	mp.Ip = net.ParseIP("192.0.2.65")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(address drop)", rc, w.Msg, dns.RcodeSuccess,
			"<nil> MsgHdr",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	mp.Ip = net.ParseIP("192.0.2.53")

	m = testutil.MakeTestDNSMsg("example.missing", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err == nil {
		t.Errorf("expected errInvalidAction but got rc: %d, msg:\n%q", rc, w.Msg)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else if err != errInvalidAction {
		t.Errorf("expected errInvalidAction but got %T: %s", err, err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	mp.Ip = net.ParseIP("192.0.2.81")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err == nil {
		t.Errorf("expected errInvalidAction but got rc: %d, msg:\n%q", rc, w.Msg)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else if err != errInvalidAction {
		t.Errorf("expected errInvalidAction but got %T: %s", err, err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}

	mp.Err = fmt.Errorf("test next plugin error")

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != mp.Err {
		t.Errorf("expected %q but got %q", mp.Err, err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}
	if !testutil.AssertDNSMessage(t, "ServeDNS(next plugin error)", rc, w.Msg, dns.RcodeSuccess,
		";; opcode: QUERY, status: SERVFAIL, id: 0\n"+
			";; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
			";; QUESTION SECTION:\n;example.com.\tIN\t A\n",
	) {
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}

	m = testutil.MakeTestDNSMsg("example.com.debug.local", dns.TypeTXT, dns.ClassCHAOS)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(next plugin error with debug)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.debug.local.\tCH\t TXT\n\n"+
				";; ANSWER SECTION:\n"+
				"example.com.debug.local.\t0\tCH\tTXT\t\"Ident: <DEBUG>\" "+
				"\"PDP response {Effect: Permit, Obligations: [policy_action: allow]}\" "+
				"\"Domain resolution: failed\"\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	mp.Err = nil
	mp.Ip = nil

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(dropped by resolver)", rc, w.Msg, dns.RcodeSuccess,
			"<nil> MsgHdr",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	mp.Ip = net.ParseIP("192.0.2.53")
	mp.Rc = dns.RcodeServerFailure

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(resolver failed)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: SERVFAIL, id: 0\n"+
				";; flags: qr aa; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	mp.Rc = dns.RcodeSuccess

	client := p.pdp

	dnErr := fmt.Errorf("test error on domain validation")
	errPep := testutil.NewErraticPep(client, dnErr, nil)
	p.pdp = errPep

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != dnErr {
		t.Errorf("expected %q but got %q", dnErr, err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}
	if !testutil.AssertDNSMessage(t, "ServeDNS(error on domain validation)", rc, w.Msg, dns.RcodeSuccess,
		"<nil> MsgHdr",
	) {
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}

	ipErr := fmt.Errorf("test error on address validation")
	errPep = testutil.NewErraticPep(client, nil, ipErr)
	p.pdp = errPep

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != ipErr {
		t.Errorf("expected %q but got %q", ipErr, err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}
	if !testutil.AssertDNSMessage(t, "ServeDNS(error on domain validation)", rc, w.Msg, dns.RcodeSuccess,
		"<nil> MsgHdr",
	) {
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}
}

func TestPolicyPluginServeDNSPassthrough(t *testing.T) {
	endpoint := "127.0.0.1:5555"
	if err := testutil.WaitForPortClosed(endpoint); err != nil {
		t.Fatalf("port still in use: %s", err)
	}

	p := newPolicyPlugin()

	p.conf.endpoints = []string{endpoint}
	p.conf.connTimeout = time.Second
	p.conf.streams = 1
	p.conf.log = true
	p.conf.passthrough = []string{"passthrough.local."}
	p.conf.debugID = "<DEBUG>"
	p.conf.debugSuffix = "debug.local."

	p.next = &testutil.MockPlugin{
		Ip: net.ParseIP("192.0.2.53"),
		Rc: dns.RcodeSuccess,
	}

	g := testutil.NewLogGrabber()
	if err := p.connect(); err != nil {
		logs := g.Release()
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		t.Fatal(err)
	}
	defer p.closeConn()

	m := testutil.MakeTestDNSMsg("example.passthrough.local", dns.TypeA, dns.ClassINET)
	w := testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err := p.ServeDNS(context.TODO(), w, m)
	logs := g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(passthrough)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr aa; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.passthrough.local.\tIN\t A\n\n"+
				";; ANSWER SECTION:\n"+
				"example.passthrough.local.\t0\tIN\tA\t192.0.2.53\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	m = testutil.MakeTestDNSMsg("example.passthrough.local.debug.local.", dns.TypeTXT, dns.ClassCHAOS)
	w = testutil.NewTestAddressedNonwriter("192.0.2.1")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), w, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS(passthrough+debug)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr aa; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.passthrough.local.debug.local.\tCH\t TXT\n\n"+
				";; ANSWER SECTION:\n"+
				"example.passthrough.local.debug.local.\t0\tCH\tTXT\t\"Ident: <DEBUG>\" \"Passthrough: yes\"\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}
}

func TestPolicyPluginServeDNSWithDnstap(t *testing.T) {
	endpoint := "127.0.0.1:5555"
	srv := testutil.StartPDPServer(t, serveDNSTestPolicy, endpoint)
	defer func() {
		if logs := srv.Stop(); len(logs) > 0 {
			t.Logf("server logs:\n%s", logs)
		}
	}()

	if err := testutil.WaitForPortOpened(endpoint); err != nil {
		t.Fatalf("can't connect to PDP server: %s", err)
	}

	p := newPolicyPlugin()
	p.conf.endpoints = []string{endpoint}
	p.conf.connTimeout = time.Second
	p.conf.streams = 1
	p.conf.log = true
	p.conf.debugID = "<DEBUG>"
	p.conf.debugSuffix = "debug.local."
	p.conf.autoResAttrs = true
	p.conf.attrs.parseAttrList(attrListTypeVal1, attrNameDomainName, `type="query"`)
	p.conf.attrs.parseAttrList(attrListTypeVal2, attrNameAddress, `type="response"`)
	p.conf.attrs.parseAttrList(attrListTypeDnstap,
		attrNameSourceIP, attrNamePolicyAction, attrNameRedirectTo)
	p.conf.attrs.parseAttrList(attrListTypeDnstap+1,
		attrNameDomainName, attrNamePolicyAction, attrNameRedirectTo)

	mp := &testutil.MockPlugin{
		Ip: net.ParseIP("192.0.2.53"),
		Rc: dns.RcodeSuccess,
	}
	p.next = mp

	io := testutil.NewIORoutine()
	p.tapIO = newPolicyDnstapSender(io)

	g := testutil.NewLogGrabber()
	if err := p.connect(); err != nil {
		logs := g.Release()
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		t.Fatal(err)
	}
	defer p.closeConn()

	m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w := testutil.NewTestAddressedNonwriterWithAddr(&net.UDPAddr{
		IP:   net.ParseIP("10.240.0.1"),
		Port: 40212,
		Zone: "",
	})

	tapRW := &taprw.ResponseWriter{
		Query:          new(dns.Msg),
		ResponseWriter: w,
		Tapper:         &dtest.TrapTapper{Full: true},
		Send:           &taprw.SendOption{Cq: false, Cr: false},
	}

	g = testutil.NewLogGrabber()
	rc, err := p.ServeDNS(context.TODO(), tapRW, m)
	logs := g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertDNSMessage(t, "ServeDNS", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr aa; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n\n"+
				";; ANSWER SECTION:\n"+
				"example.com.\t0\tIN\tA\t192.0.2.53\n",
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}

		if !testutil.AssertCRExtraResult(t, "sendCRExtraMsg(actionAllow)", io, w.Msg,
			&pb.DnstapAttribute{Id: attrNameSourceIP, Value: "10.240.0.1"},
			&pb.DnstapAttribute{Id: attrNamePolicyAction, Value: "2"},
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	m = testutil.MakeTestDNSMsg("example.log", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriterWithAddr(&net.UDPAddr{
		IP:   net.ParseIP("10.240.0.1"),
		Port: 40212,
		Zone: "",
	})

	tapRW = &taprw.ResponseWriter{
		Query:          new(dns.Msg),
		ResponseWriter: w,
		Tapper:         &dtest.TrapTapper{Full: true},
		Send:           &taprw.SendOption{Cq: false, Cr: false},
	}

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), tapRW, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertCRExtraResult(t, "sendCRExtraMsg(actionLog0)", io, w.Msg,
			&pb.DnstapAttribute{Id: attrNameSourceIP, Value: "10.240.0.1"},
			&pb.DnstapAttribute{Id: attrNamePolicyAction, Value: "4"},
			&pb.DnstapAttribute{Id: attrNameRedirectTo, Value: "redir.ect"},
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}

	m = testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = testutil.NewTestAddressedNonwriterWithAddr(&net.UDPAddr{
		IP:   net.ParseIP("10.240.0.1"),
		Port: 40212,
		Zone: "",
	})

	tapRW = &taprw.ResponseWriter{
		Query:          new(dns.Msg),
		ResponseWriter: w,
		Tapper:         &dtest.TrapTapper{Full: true},
		Send:           &taprw.SendOption{Cq: false, Cr: false},
	}

	mp.Ip = net.ParseIP("192.0.2.97")

	g = testutil.NewLogGrabber()
	rc, err = p.ServeDNS(context.TODO(), tapRW, m)
	logs = g.Release()
	if err != nil {
		t.Error(err)
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	} else {
		if !testutil.AssertCRExtraResult(t, "sendCRExtraMsg(actionLog1)", io, w.Msg,
			&pb.DnstapAttribute{Id: attrNameDomainName, Value: "example.com."},
			&pb.DnstapAttribute{Id: attrNamePolicyAction, Value: "3"},
		) {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}
}

const serveDNSTestPolicy = `# All Permit Policy
attributes:
  type: string
  domain_name: domain
  address: address
  redirect_to: string
  policy_action: integer
  log: integer
  missing: string
policies:
  alg: FirstApplicableEffect
  policies:
  - id: "Query rules"
    target:
    - equal:
      - attr: type
      - val:
          type: string
          content: query
    alg: FirstApplicableEffect
    rules:
    - id: "Permit example.com"
      target:
      - contains:
        - val:
            type: set of domains
            content:
            - example.com
        - attr: domain_name
      effect: Permit
      obligations:
      - policy_action:
          val:
            type: integer
            content: 2
    - id: "Redirect example.log and log 0"
      target:
      - contains:
        - val:
            type: set of domains
            content:
            - example.log
        - attr: domain_name
      effect: Deny
      obligations:
      - policy_action:
          val:
            type: integer
            content: 4
      - redirect_to:
          val:
            type: string
            content: redir.ect
      - log:
          val:
            type: integer
            content: 0
    - id: "Redirect example.redirect"
      target:
      - contains:
        - val:
            type: set of domains
            content:
            - example.redirect
        - attr: domain_name
      effect: Deny
      obligations:
      - policy_action:
          val:
            type: integer
            content: 4
      - redirect_to:
          val:
            type: string
            content: "192.0.2.254"
    - id: "Block example.block"
      target:
      - contains:
        - val:
            type: set of domains
            content:
            - example.block
        - attr: domain_name
      effect: Deny
      obligations:
      - policy_action:
          val:
            type: integer
            content: 3
    - id: "Refuse example.refuse"
      target:
      - contains:
        - val:
            type: set of domains
            content:
            - example.refuse
        - attr: domain_name
      effect: Deny
      obligations:
      - policy_action:
          val:
            type: integer
            content: 5
    - id: "Drop example.drop"
      target:
      - contains:
        - val:
            type: set of domains
            content:
            - example.drop
        - attr: domain_name
      effect: Deny
      obligations:
      - policy_action:
          val:
            type: integer
            content: 1
    - id: "Missing attribute example.missing"
      target:
      - contains:
        - val:
            type: set of domains
            content:
            - example.missing
        - attr: domain_name
      condition:
        equal:
        - attr: missing
        - val:
            type: string
            content: missing
      effect: Permit
  - id: "Response rules"
    target:
    - equal:
      - attr: type
      - val:
          type: string
          content: response
    alg: FirstApplicableEffect
    rules:
    - id: "Permit 192.0.2.48/28"
      target:
      - contains:
        - val:
            type: set of networks
            content:
            - 192.0.2.48/28
        - attr: address
      effect: Permit
      obligations:
      - policy_action:
          val:
            type: integer
            content: 2
    - id: "Block 192.0.2.96/28 and log 1"
      target:
      - contains:
        - val:
            type: set of networks
            content:
            - 192.0.2.96/28
        - attr: address
      effect: Permit
      obligations:
      - policy_action:
          val:
            type: integer
            content: 3
      - log:
          val:
            type: integer
            content: 1
    - id: "Redirect 192.0.2.0/28"
      target:
      - contains:
        - val:
            type: set of networks
            content:
            - 192.0.2.0/28
        - attr: address
      effect: Deny
      obligations:
      - policy_action:
          val:
            type: integer
            content: 4
      - redirect_to:
          val:
            type: string
            content: "192.0.2.253"
    - id: "Block 192.0.2.16/28"
      target:
      - contains:
        - val:
            type: set of networks
            content:
            - 192.0.2.16/28
        - attr: address
      effect: Deny
      obligations:
      - policy_action:
          val:
            type: integer
            content: 3
    - id: "Refuse 192.0.2.32/28"
      target:
      - contains:
        - val:
            type: set of networks
            content:
            - 192.0.2.32/28
        - attr: address
      effect: Deny
      obligations:
      - policy_action:
          val:
            type: integer
            content: 5
    - id: "Drop 192.0.2.64/28"
      target:
      - contains:
        - val:
            type: set of networks
            content:
            - 192.0.2.64/28
        - attr: address
      effect: Deny
      obligations:
      - policy_action:
          val:
            type: integer
            content: 1
    - id: "Missing attribute 192.0.2.80/28"
      target:
      - contains:
        - val:
            type: set of networks
            content:
            - 192.0.2.80/28
        - attr: address
      condition:
        equal:
        - attr: missing
        - val:
            type: string
            content: missing
      effect: Permit
`
