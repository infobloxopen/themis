package pep

import (
	"fmt"
	"strings"
	"testing"

	"github.com/infobloxopen/themis/pdpserver/server"
)

const allPermitPolicy = `# Policy for client tests
attributes:
  x: string

policies:
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
    obligations:
    - x:
       val:
         type: string
         content: AllPermitRule
`

func TestUnaryClientValidation(t *testing.T) {
	pdp := startTestPDPServer(allPermitPolicy, 5555, t)
	defer func() {
		if logs := pdp.Stop(); len(logs) > 0 {
			t.Logf("server logs:\n%s", logs)
		}
	}()

	c := NewClient()
	err := c.Connect("127.0.0.1:5555")
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
	defer c.Close()

	in := decisionRequest{
		Direction: "Any",
		Policy:    "AllPermitPolicy",
		Domain:    "example.com",
	}
	var out decisionResponse
	err = c.Validate(in, &out)
	if err != nil {
		t.Errorf("expected no error but got %s", err)
	}

	if out.Effect != "PERMIT" || out.Reason != "Ok" || out.X != "AllPermitRule" {
		t.Errorf("got unexpected response: %#v", out)
	}
}

func startTestPDPServer(p string, s uint16, t *testing.T) *loggedServer {
	service := fmt.Sprintf(":%d", s)
	primary := newServer(
		server.WithServiceAt(service),
	)
	addr := "127.0.0.1" + service

	if err := primary.s.ReadPolicies(strings.NewReader(p)); err != nil {
		t.Fatalf("can't read policies: %s", err)
	}

	if err := waitForPortClosed(addr); err != nil {
		t.Fatalf("port still in use: %s", err)
	}
	go func() {
		if err := primary.s.Serve(); err != nil {
			t.Fatalf("server failed: %s", err)
		}
	}()

	if err := waitForPortOpened(addr); err != nil {
		if logs := primary.Stop(); len(logs) > 0 {
			t.Logf("server logs:\n%s", logs)
		}

		t.Fatalf("can't connect to PDP server: %s", err)
	}
	return primary
}
