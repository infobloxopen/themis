package policy

import (
	"net"
	"testing"
	"time"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
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

func TestPolicyPluginServeDNS(t *testing.T) {
	endpoint := "127.0.0.1:5555"
	srv := startPDPServer(t, serveDNSTestPolicy, endpoint)
	defer func() {
		if logs := srv.Stop(); len(logs) > 0 {
			t.Logf("server logs:\n%s", logs)
		}
	}()

	if err := waitForPortOpened(endpoint); err != nil {
		t.Fatalf("can't connect to PDP server: %s", err)
	}

	p := newPolicyPlugin()
	p.conf.endpoints = []string{endpoint}
	p.conf.connTimeout = time.Second
	p.conf.streams = 1
	p.conf.log = true
	p.conf.debugID = "<DEBUG>"
	p.conf.debugSuffix = "debug.infoblox.com."

	mp := &mockPlugin{
		ip: net.ParseIP("192.0.2.53"),
	}
	p.next = mp

	if err := p.connect(); err != nil {
		t.Fatal(err)
	}
	defer p.closeConn()

	m := makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w := newTestAddressedNonwriter("192.0.2.1")

	rc, err := p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr aa; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n\n"+
				";; ANSWER SECTION:\n"+
				"example.com.\t0\tIN\tA\t192.0.2.53\n",
		)
	}

	m = makeTestDNSMsg("example.com.debug.infoblox.com", dns.TypeTXT, dns.ClassCHAOS)
	w = newTestAddressedNonwriter("192.0.2.1")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(debug)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.debug.infoblox.com.\tCH\t TXT\n\n"+
				";; ANSWER SECTION:\n"+
				"example.com.debug.infoblox.com.\t0\tCH\tTXT\t\"resolve:yes,query:'allow',ident:'<DEBUG>'\"\n",
		)
	}

	m = makeTestDNSMsg("example.redirect", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(domain redirect)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.redirect.\tIN\t A\n\n"+
				";; ANSWER SECTION:\n"+
				"example.redirect.\t0\tIN\tA\t192.0.2.254\n",
		)
	}

	m = makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	mp.ip = net.ParseIP("192.0.2.1")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(address redirect)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n\n"+
				";; ANSWER SECTION:\n"+
				"example.com.\t0\tIN\tA\t192.0.2.253\n",
		)
	}

	mp.ip = net.ParseIP("192.0.2.53")

	m = makeTestDNSMsg("example.block", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(domain block)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NXDOMAIN, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.block.\tIN\t A\n",
		)
	}

	m = makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	mp.ip = net.ParseIP("192.0.2.17")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(address block)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NXDOMAIN, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n",
		)
	}

	mp.ip = net.ParseIP("192.0.2.53")

	m = makeTestDNSMsg("example.refuse", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(domain refuse)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: REFUSED, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.refuse.\tIN\t A\n",
		)
	}

	m = makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	mp.ip = net.ParseIP("192.0.2.33")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(address refuse)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: REFUSED, id: 0\n"+
				";; flags: qr; QUERY: 1, ANSWER: 0, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.com.\tIN\t A\n",
		)
	}

	mp.ip = net.ParseIP("192.0.2.53")

	m = makeTestDNSMsg("example.drop", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(domain drop)", rc, w.Msg, dns.RcodeSuccess,
			"<nil> MsgHdr",
		)
	}

	m = makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	mp.ip = net.ParseIP("192.0.2.65")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(address drop)", rc, w.Msg, dns.RcodeSuccess,
			"<nil> MsgHdr",
		)
	}

	mp.ip = net.ParseIP("192.0.2.53")

	m = makeTestDNSMsg("example.missing", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err == nil {
		t.Errorf("expected errInvalidAction but got rc: %d, msg:\n%q", rc, w.Msg)
	} else if err != errInvalidAction {
		t.Errorf("exepcted errInvalidAction but got %T: %s", err, err)
	}

	m = makeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
	w = newTestAddressedNonwriter("192.0.2.1")

	mp.ip = net.ParseIP("192.0.2.81")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err == nil {
		t.Errorf("expected errInvalidAction but got rc: %d, msg:\n%q", rc, w.Msg)
	} else if err != errInvalidAction {
		t.Errorf("exepcted errInvalidAction but got %T: %s", err, err)
	}
}

func TestPolicyPluginServeDNSPassthrough(t *testing.T) {
	endpoint := "127.0.0.1:5555"
	if err := waitForPortClosed(endpoint); err != nil {
		t.Fatalf("port still in use: %s", err)
	}

	p := newPolicyPlugin()

	p.conf.endpoints = []string{endpoint}
	p.conf.connTimeout = time.Second
	p.conf.streams = 1
	p.conf.log = true
	p.conf.passthrough = []string{"passthrough.infoblox.com."}
	p.conf.debugID = "<DEBUG>"
	p.conf.debugSuffix = "debug.infoblox.com."

	p.next = &mockPlugin{
		ip: net.ParseIP("192.0.2.53"),
	}

	if err := p.connect(); err != nil {
		t.Fatal(err)
	}
	defer p.closeConn()

	m := makeTestDNSMsg("example.passthrough.infoblox.com", dns.TypeA, dns.ClassINET)
	w := newTestAddressedNonwriter("192.0.2.1")

	rc, err := p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(passthrough)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr aa; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.passthrough.infoblox.com.\tIN\t A\n\n"+
				";; ANSWER SECTION:\n"+
				"example.passthrough.infoblox.com.\t0\tIN\tA\t192.0.2.53\n",
		)
	}

	m = makeTestDNSMsg("example.passthrough.infoblox.com.debug.infoblox.com.", dns.TypeTXT, dns.ClassCHAOS)
	w = newTestAddressedNonwriter("192.0.2.1")

	rc, err = p.ServeDNS(context.TODO(), w, m)
	if err != nil {
		t.Error(err)
	} else {
		assertDNSMessage(t, "ServeDNS(passthrough+debug)", rc, w.Msg, dns.RcodeSuccess,
			";; opcode: QUERY, status: NOERROR, id: 0\n"+
				";; flags: qr aa; QUERY: 1, ANSWER: 1, AUTHORITY: 0, ADDITIONAL: 0\n\n"+
				";; QUESTION SECTION:\n"+
				";example.passthrough.infoblox.com.debug.infoblox.com.\tCH\t TXT\n\n"+
				";; ANSWER SECTION:\n"+
				"example.passthrough.infoblox.com.debug.infoblox.com.\t0\tCH\tTXT\t\"action:passthrough\"\n",
		)
	}
}

const serveDNSTestPolicy = `# All Permit Policy
attributes:
  type: string
  domain_name: domain
  address: address
  redirect_to: string
  refuse: string
  drop: string
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
      - refuse:
          val:
            type: string
            content: ""
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
      - drop:
          val:
            type: string
            content: ""
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
      - refuse:
          val:
            type: string
            content: ""
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
      - drop:
          val:
            type: string
            content: ""
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

func assertDNSMessage(t *testing.T, desc string, rc int, m *dns.Msg, erc int, eMsg string) {
	if rc != erc {
		t.Errorf("expected %d rcode but got %d", erc, rc)
	}

	if m.String() != eMsg {
		t.Errorf("expected response:\n%q\nbut got:\n%q", eMsg, m)
	}
}

type mockPlugin struct {
	ip net.IP
}

// Name implements the plugin.Handler interface.
func (p *mockPlugin) Name() string {
	return "mockPlugin"
}

// ServeDNS implements the plugin.Handler interface.
func (p *mockPlugin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	if r == nil || len(r.Question) <= 0 {
		return dns.RcodeServerFailure, nil
	}

	q := r.Question[0]
	hdr := dns.RR_Header{
		Name:   q.Name,
		Rrtype: q.Qtype,
		Class:  q.Qclass,
	}

	if ipv4 := p.ip.To4(); ipv4 != nil {
		if q.Qtype != dns.TypeA {
			return dns.RcodeSuccess, nil
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true

		m.Answer = append(m.Answer,
			&dns.A{
				Hdr: hdr,
				A:   ipv4,
			},
		)

		w.WriteMsg(m)
	} else if ipv6 := p.ip.To16(); ipv6 != nil {
		if q.Qtype != dns.TypeAAAA {
			return dns.RcodeSuccess, nil
		}

		m := new(dns.Msg)
		m.SetReply(r)
		m.Authoritative = true

		m.Answer = append(m.Answer,
			&dns.AAAA{
				Hdr:  hdr,
				AAAA: ipv6,
			},
		)

		w.WriteMsg(m)
	}

	return dns.RcodeSuccess, nil
}
