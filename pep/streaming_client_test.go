package pep

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestStreamingClientValidation(t *testing.T) {
	pdp := startTestPDPServer(allPermitPolicy, 5555, t)
	defer func() {
		if logs := pdp.Stop(); len(logs) > 0 {
			t.Logf("server logs:\n%s", logs)
		}
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
	firstPDP := startTestPDPServer(allPermitPolicy, 5555, t)
	defer func() {
		if logs := firstPDP.Stop(); len(logs) > 0 {
			t.Logf("primary server logs:\n%s", logs)
		}
	}()

	secondPDP := startTestPDPServer(allPermitPolicy, 5556, t)
	defer func() {
		if logs := secondPDP.Stop(); len(logs) > 0 {
			t.Logf("secondary server logs:\n%s", logs)
		}
	}()

	c := NewClient(
		WithStreams(2),
		WithRoundRobinBalancer(
			"127.0.0.1:5555",
			"127.0.0.1:5556",
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
	firstPDP := startTestPDPServer(allPermitPolicy, 5555, t)
	defer func() {
		if logs := firstPDP.Stop(); len(logs) > 0 {
			t.Logf("primary server logs:\n%s", logs)
		}
	}()

	secondPDP := startTestPDPServer(allPermitPolicy, 5556, t)
	defer func() {
		if logs := secondPDP.Stop(); len(logs) > 0 {
			t.Logf("secondary server logs:\n%s", logs)
		}
	}()

	c := NewClient(
		WithStreams(2),
		WithHotSpotBalancer(
			"127.0.0.1:5555",
			"127.0.0.1:5556",
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

func TestStreamingClientValidationNoConnection(t *testing.T) {
	c := NewClient(WithStreams(1))
	err := c.Connect("127.0.0.1:5555")
	if err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
	defer c.Close()

	done := make(chan bool)

	go func() {
		in := decisionRequest{
			Direction: "Any",
			Policy:    "AllPermitPolicy",
			Domain:    "example.com",
		}
		var out decisionResponse
		err = c.Validate(in, &out)
		if err != nil {
			if err != ErrorNotConnected {
				t.Errorf("expected not connected error but got %s", err)
			}
		} else {
			t.Errorf("expected error but got response: %#v", out)
		}

		close(done)
	}()

	select {
	case <-time.After(10 * time.Second):
		t.Errorf("expected no connection error but got nothing after 10 seconds")
		c.Close()

	case <-done:
	}
}
