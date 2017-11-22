package pep

import (
	"testing"

	pb "github.com/infobloxopen/themis/pdp-service"
)

const (
	fakeServerAddress    = "localhost:5555"
	fakeServerAltAddress = "localhost:5556"
)

func TestStreamClientRecovery(t *testing.T) {
	s, err := newFailServer(fakeServerAddress)
	if err != nil {
		t.Fatalf("couldn't start fake server: %s", err)
	}

	defer s.Stop()

	err = waitForPortOpened(fakeServerAddress)
	if err != nil {
		t.Fatalf("can't connect to fake server: %s", err)
	}

	c := NewClient(WithStreams(1))
	err = c.Connect(fakeServerAddress)
	if err != nil {
		t.Fatalf("can't connect to fake server: %s", err)
	}

	in := pb.Request{
		Attributes: []*pb.Attribute{
			{
				Id:    IDID,
				Value: "1",
			},
			{
				Id:    failID,
				Value: thisRequest,
			},
		},
	}

	var out pb.Request
	err = c.Validate(in, &out)
	if err != nil {
		t.Fatalf("can't send first request: %s", err)
	}

	var attempts uint64 = 2
	if s.ID != attempts {
		t.Errorf("Expected %d attempts but got %d", attempts, s.ID)
	}
}

func TestStreamClientRecoveryWithHotSpotBalancer(t *testing.T) {
	s1, err := newFailServer(fakeServerAddress)
	if err != nil {
		t.Fatalf("couldn't start fake server: %s", err)
	}

	defer s1.Stop()

	err = waitForPortOpened(fakeServerAddress)
	if err != nil {
		t.Fatalf("can't connect to fake server: %s", err)
	}

	s2, err := newFailServer(fakeServerAltAddress)
	if err != nil {
		t.Fatalf("couldn't start fake server: %s", err)
	}

	defer s2.Stop()

	err = waitForPortOpened(fakeServerAltAddress)
	if err != nil {
		t.Fatalf("can't connect to fake server: %s", err)
	}

	c := NewClient(
		WithStreams(1),
		WithHotSpotBalancer(
			fakeServerAddress,
			fakeServerAltAddress,
		),
	)
	err = c.Connect(fakeServerAddress)
	if err != nil {
		t.Fatalf("can't connect to fake server: %s", err)
	}

	in := pb.Request{
		Attributes: []*pb.Attribute{
			{
				Id:    IDID,
				Value: "1",
			},
			{
				Id:    failID,
				Value: thisRequest,
			},
		},
	}

	var out pb.Request
	err = c.Validate(in, &out)
	if err != nil {
		t.Fatalf("can't send first request: %s", err)
	}

	var attempts uint64 = 2
	total := s1.ID + s2.ID
	if total != attempts {
		t.Errorf("Expected %d attempts but got %d", attempts, total)
	}
}
