package policy

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	"github.com/infobloxopen/themis/pdp"
	_ "github.com/infobloxopen/themis/pdp/selector"
	"github.com/miekg/dns"
)

const testPolicy = `# Policy set for client interaction tests
attributes:
  type: string
  domain_name: domain
  address: address
  rule: string
  log: string
  o1: string
  o2: string
  o3: string

policies:
  alg: FirstApplicableEffect
  policies:
  - id: "Query Policy"
    target:
    - equal:
      - attr: type
      - val:
          type: string
          content: query
    alg: FirstApplicableEffect
    rules:
    - id: "Query for example.com"
      target:
      - contains:
        - val:
            type: set of domains
            content:
            - example.com
        - attr: domain_name
      effect: Permit
      obligations:
      - rule:
          val:
            type: string
            content: "Query rule for example.com"
    - id: "Many obligations rule"
      target:
      - contains:
        - val:
            type: set of domains
            content:
            - overflow.me
        - attr: domain_name
      effect: Permit
      obligations:
      - rule:
          val:
            type: string
            content: "Many obligations rule"
      - o1:
          val:
            type: string
            content: "First additional obligation"
      - o2:
          val:
            type: string
            content: "Second additional obligation"
      - o3:
          val:
            type: string
            content: "Third additional obligation"
  - id: "Response Policy"
    target:
    - equal:
      - attr: type
      - val:
          type: string
          content: response
    alg: FirstApplicableEffect
    rules:
    - id: "Response for 192.0.2.0/28"
      target:
      - contains:
        - val:
            type: set of networks
            content:
            - 192.0.2.0/28
        - attr: address
      effect: Permit
      obligations:
      - rule:
          val:
            type: string
            content: "Response rule for 192.0.2.0/28"
      - log:
          val:
            type: string
            content: ""
`

func TestStreamingClientInteraction(t *testing.T) {
	endpoint := "127.0.0.1:5555"
	srv := testutil.StartPDPServer(t, testPolicy, endpoint)
	defer func() {
		if logs := srv.Stop(); len(logs) > 0 {
			t.Logf("server logs:\n%s", logs)
		}
	}()

	if err := testutil.WaitForPortOpened(endpoint); err != nil {
		t.Fatalf("can't connect to PDP server: %s", err)
	}

	g := testutil.NewLogGrabber()
	ok := t.Run("noCache", func(t *testing.T) {
		p := newPolicyPlugin()
		p.conf.endpoints = []string{endpoint}
		p.conf.connTimeout = time.Second
		p.conf.streams = 1
		p.conf.log = true

		if err := p.connect(); err != nil {
			t.Fatal(err)
		}
		defer p.closeConn()

		m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
		w := testutil.NewTestAddressedNonwriter("192.0.2.1")

		ah := newAttrHolderWithDnReq(w, m, p.conf.options, nil)
		attrs := make([]pdp.AttributeAssignment, p.conf.maxResAttrs)
		if err := p.validate(ah, attrs); err != nil {
			t.Error(err)
		}

		if ah.action != actionAllow {
			aName := fmt.Sprintf("unknown action %d", ah.action)
			if ah.action >= 0 && int(ah.action) < len(actionNames) {
				aName = actionNames[ah.action]
			}
			t.Errorf("expected %q action but got %q", actionNames[actionAllow], aName)
		}

		pdp.AssertAttributeAssignments(t, "p.validate(domain request)", ah.dnRes,
			pdp.MakeStringAssignment("rule", "Query rule for example.com"),
		)

		ah.addIPReq(net.ParseIP("192.0.2.1"))

		attrs = make([]pdp.AttributeAssignment, p.conf.maxResAttrs)
		if err := p.validate(ah, attrs); err != nil {
			t.Error(err)
		}

		if ah.action != actionLog {
			aName := fmt.Sprintf("unknown action %d", ah.action)
			if ah.action >= 0 && int(ah.action) < len(actionNames) {
				aName = actionNames[ah.action]
			}
			t.Errorf("expected %q action but got %q", actionNames[actionLog], aName)
		}

		pdp.AssertAttributeAssignments(t, "p.validate(domain request)", ah.ipRes,
			pdp.MakeStringAssignment("rule", "Response rule for 192.0.2.0/28"),
			pdp.MakeStringAssignment("log", ""),
		)
	})

	logs := g.Release()
	if !ok {
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}

	g = testutil.NewLogGrabber()
	ok = t.Run("cacheTTL", func(t *testing.T) {
		p := newPolicyPlugin()
		p.conf.endpoints = []string{endpoint}
		p.conf.connTimeout = time.Second
		p.conf.streams = 1
		p.conf.log = true
		p.conf.maxReqSize = 128
		p.conf.cacheTTL = 10 * time.Minute

		if err := p.connect(); err != nil {
			t.Fatal(err)
		}
		defer p.closeConn()

		m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
		w := testutil.NewTestAddressedNonwriter("192.0.2.1")

		ah := newAttrHolderWithDnReq(w, m, p.conf.options, nil)
		attrs := make([]pdp.AttributeAssignment, p.conf.maxResAttrs)
		if err := p.validate(ah, attrs); err != nil {
			t.Error(err)
		}

		if ah.action != actionAllow {
			aName := fmt.Sprintf("unknown action %d", ah.action)
			if ah.action >= 0 && int(ah.action) < len(actionNames) {
				aName = actionNames[ah.action]
			}
			t.Errorf("expected %q action but got %q", actionNames[actionAllow], aName)
		}

		pdp.AssertAttributeAssignments(t, "p.validate(domain request)", ah.dnRes,
			pdp.MakeStringAssignment("rule", "Query rule for example.com"),
		)
	})

	logs = g.Release()
	if !ok {
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}

	g = testutil.NewLogGrabber()
	ok = t.Run("cacheTTLAndLimit", func(t *testing.T) {
		p := newPolicyPlugin()
		p.conf.endpoints = []string{endpoint}
		p.conf.connTimeout = time.Second
		p.conf.streams = 1
		p.conf.log = true
		p.conf.maxReqSize = 128
		p.conf.cacheTTL = 10 * time.Minute
		p.conf.cacheLimit = 128

		if err := p.connect(); err != nil {
			t.Fatal(err)
		}
		defer p.closeConn()

		m := testutil.MakeTestDNSMsg("example.com", dns.TypeA, dns.ClassINET)
		w := testutil.NewTestAddressedNonwriter("192.0.2.1")

		ah := newAttrHolderWithDnReq(w, m, p.conf.options, nil)
		attrs := make([]pdp.AttributeAssignment, p.conf.maxResAttrs)
		if err := p.validate(ah, attrs); err != nil {
			t.Error(err)
		}

		if ah.action != actionAllow {
			aName := fmt.Sprintf("unknown action %d", ah.action)
			if ah.action >= 0 && int(ah.action) < len(actionNames) {
				aName = actionNames[ah.action]
			}
			t.Errorf("expected %q action but got %q", actionNames[actionAllow], aName)
		}

		pdp.AssertAttributeAssignments(t, "p.validate(domain request)", ah.dnRes,
			pdp.MakeStringAssignment("rule", "Query rule for example.com"),
		)
	})

	logs = g.Release()
	if !ok {
		t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
	}
}

func TestStreamingClientInteractionWithObligationsOverflow(t *testing.T) {
	endpoint := "127.0.0.1:5555"
	srv := testutil.StartPDPServer(t, testPolicy, endpoint)
	defer func() {
		if logs := srv.Stop(); len(logs) > 0 {
			t.Logf("server logs:\n%s", logs)
		}
	}()

	if err := testutil.WaitForPortOpened(endpoint); err != nil {
		t.Fatalf("can't connect to PDP server: %s", err)
	}

	ok := true
	g := testutil.NewLogGrabber()
	defer func() {
		logs := g.Release()
		if !ok {
			t.Logf("=== plugin logs ===\n%s--- plugin logs ---", logs)
		}
	}()

	p := newPolicyPlugin()
	p.conf.endpoints = []string{endpoint}
	p.conf.connTimeout = time.Second
	p.conf.streams = 1
	p.conf.maxResAttrs = 3
	p.conf.log = true

	if err := p.connect(); err != nil {
		t.Fatal(err)
		ok = false
	}
	defer p.closeConn()

	m := testutil.MakeTestDNSMsg("overflow.me", dns.TypeA, dns.ClassINET)
	w := testutil.NewTestAddressedNonwriter("192.0.2.1")

	ah := newAttrHolderWithDnReq(w, m, p.conf.options, nil)
	attrs := make([]pdp.AttributeAssignment, p.conf.maxResAttrs)
	err := p.validate(ah, attrs)
	if err == nil {
		aName := fmt.Sprintf("unknown action %d", ah.action)
		if ah.action >= 0 && int(ah.action) < len(actionNames) {
			aName = actionNames[ah.action]
		}

		t.Errorf("expected response overflow error but got %q response:\n:%+v", aName, ah.dnRes)
		ok = false
	}
}
