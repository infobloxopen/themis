package pep

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestStreamingClientValidation(t *testing.T) {
	tmpYAST, pdp := startTestPDPServer(allPermitPolicy, "127.0.0.1:5555", "127.0.0.1:5554", t)
	defer func() {
		_, errDump, _ := pdp.kill()
		if t.Failed() && len(errDump) > 0 {
			t.Logf("PDP server dump:\n%s", strings.Join(errDump, "\n"))
		}

		os.Remove(tmpYAST)
	}()

	c := NewClient(WithStreams(1))
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

func TestStreamingClientValidationWithRoundRobingBalancer(t *testing.T) {
	firstTmpYAST, firstPDP := startTestPDPServer(allPermitPolicy, "127.0.0.1:5515", "127.0.0.1:5514", t)
	defer func() {
		_, errDump, _ := firstPDP.kill()
		if t.Failed() && len(errDump) > 0 {
			t.Logf("first PDP server dump:\n%s", strings.Join(errDump, "\n"))
		}

		os.Remove(firstTmpYAST)
	}()

	secondTmpYAST, secondPDP := startTestPDPServer(allPermitPolicy, "127.0.0.1:5525", "127.0.0.1:5524", t)
	defer func() {
		_, errDump, _ := secondPDP.kill()
		if t.Failed() && len(errDump) > 0 {
			t.Logf("second PDP server dump:\n%s", strings.Join(errDump, "\n"))
		}

		os.Remove(secondTmpYAST)
	}()

	c := NewClient(
		WithStreams(2),
		WithRoundRobinBalancer(
			"127.0.0.1:5515",
			"127.0.0.1:5525",
		))
	err := c.Connect("")
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

func TestStreamingClientValidationWithHotSpotBalancer(t *testing.T) {
	firstTmpYAST, firstPDP := startTestPDPServer(allPermitPolicy, "127.0.0.1:5515", "127.0.0.1:5514", t)
	defer func() {
		_, errDump, _ := firstPDP.kill()
		if t.Failed() && len(errDump) > 0 {
			t.Logf("first PDP server dump:\n%s", strings.Join(errDump, "\n"))
		}

		os.Remove(firstTmpYAST)
	}()

	secondTmpYAST, secondPDP := startTestPDPServer(allPermitPolicy, "127.0.0.1:5525", "127.0.0.1:5524", t)
	defer func() {
		_, errDump, _ := secondPDP.kill()
		if t.Failed() && len(errDump) > 0 {
			t.Logf("second PDP server dump:\n%s", strings.Join(errDump, "\n"))
		}

		os.Remove(secondTmpYAST)
	}()

	c := NewClient(
		WithStreams(2),
		WithHotSpotBalancer(
			"127.0.0.1:5515",
			"127.0.0.1:5525",
		))
	err := c.Connect("")
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
	defer c.Close()

	in := decisionRequest{
		Direction: "Any",
		Policy:    "AllPermitPolicy",
		Domain:    "example.com",
	}

	errs := make([]error, 10)
	var wg sync.WaitGroup
	for i := range errs {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			var out decisionResponse
			err := c.Validate(in, &out)
			if err != nil {
				errs[i] = err
			} else if out.Effect != "PERMIT" || out.Reason != "Ok" || out.X != "AllPermitRule" {
				errs[i] = fmt.Errorf("got unexpected response: %#v", out)
			}
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("requset %d failed with error %s", i, err)
		}
	}
}
