package policy

import (
	"net"
	"testing"
	"time"

	"github.com/infobloxopen/themis/contrib/coredns/policy/testutil"
	"github.com/infobloxopen/themis/pdp"
	_ "github.com/infobloxopen/themis/pdp/selector"
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

		req := []pdp.AttributeAssignment{
			pdp.MakeStringAssignment("type", "query"),
			pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeDnOrFail(t, "example.com")),
		}
		res := pdp.Response{}
		err := p.pdp.Validate(req, &res)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		testutil.AssertPdpResponse(t, &res, pdp.EffectPermit,
			pdp.MakeStringAssignment("rule", "Query rule for example.com"),
		)

		req = []pdp.AttributeAssignment{
			pdp.MakeStringAssignment("type", "response"),
			pdp.MakeAddressAssignment(attrNameAddress, net.ParseIP("192.0.2.1")),
		}
		res = pdp.Response{}
		err = p.pdp.Validate(req, &res)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		testutil.AssertPdpResponse(t, &res, pdp.EffectPermit,
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

		req := []pdp.AttributeAssignment{
			pdp.MakeStringAssignment("type", "query"),
			pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeDnOrFail(t, "example.com")),
		}
		res := pdp.Response{}
		err := p.pdp.Validate(req, &res)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		testutil.AssertPdpResponse(t, &res, pdp.EffectPermit,
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

		req := []pdp.AttributeAssignment{
			pdp.MakeStringAssignment("type", "query"),
			pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeDnOrFail(t, "example.com")),
		}
		res := pdp.Response{}
		err := p.pdp.Validate(req, &res)
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
		testutil.AssertPdpResponse(t, &res, pdp.EffectPermit,
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

	req := []pdp.AttributeAssignment{
		pdp.MakeStringAssignment("type", "query"),
		pdp.MakeDomainAssignment(attrNameDomainName, testutil.MakeDnOrFail(t, "overflow.me")),
	}
	res := pdp.Response{Obligations: make([]pdp.AttributeAssignment, 1)}
	err := p.pdp.Validate(req, &res)
	if err == nil {
		t.Errorf("expected response overflow error but got response:\n:%+v", res)
		ok = false
	}
}
