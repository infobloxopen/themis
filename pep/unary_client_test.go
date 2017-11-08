package pep

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const allPermitPolicy = `# Policy for client tests
attributes:
  x: string

policies:
  alg: FirstApplicableEffect
  rules:
  - effect: Permit
    obligations:
    - x: AllPermitRule
`

func TestUnaryClientValidation(t *testing.T) {
	tmpYAST, pdp := startTestPDPServer(allPermitPolicy, "127.0.0.1:5555", "127.0.0.1:5554", t)
	defer func() {
		_, errDump, _ := pdp.kill()
		if t.Failed() && len(errDump) > 0 {
			t.Logf("PDP server dump:\n%s", strings.Join(errDump, "\n"))
		}

		os.Remove(tmpYAST)
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

func startTestPDPServer(p, s, c string, t *testing.T) (string, *proc) {
	tmpYAST, err := makeTempFile(p, "policy")
	if err != nil {
		t.Fatalf("can't create policy file: %s", err)
	}

	binPath := filepath.Join(build.Default.GOPATH, "/src/github.com/infobloxopen/themis/build/pdpserver")
	pdp, err := newProc(binPath, "-l", s, "-c", c, "-p", tmpYAST, "-v", "3")
	if err != nil {
		os.Remove(tmpYAST)
		t.Fatalf("can't start PDP server: %s", err)
	}

	err = waitForPortOpened(s)
	if err != nil {
		_, errDump, _ := pdp.kill()
		if len(errDump) > 0 {
			t.Logf("%s PDP server dump:\n%s", s, strings.Join(errDump, "\n"))
		}
		os.Remove(tmpYAST)
		t.Fatalf("can't connect to PDP server: %s", err)
	}

	return tmpYAST, pdp
}
