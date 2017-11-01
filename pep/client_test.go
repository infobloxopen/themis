package pep

import "testing"

func TestNewClient(t *testing.T) {
	c := NewClient()
	if _, ok := c.(*pdpUnaryClient); !ok {
		t.Errorf("Expected *pdpUnaryClient from NewClient got %v", c)
	}
}

func TestNewBalancedClient(t *testing.T) {
	c := NewClient(WithBalancer("127.0.0.1:1000", "127.0.0.1:1001"))
	if uc, ok := c.(*pdpUnaryClient); ok {
		if len(uc.opts.addresses) <= 0 {
			t.Errorf("Expected balancer to be set but got nothing")
		}
	} else {
		t.Errorf("Expected *pdpUnaryClient from NewClient got %v", c)
	}
}

func TestNewStreamingClient(t *testing.T) {
	c := NewClient(WithStreams(5))
	if sc, ok := c.(*pdpStreamingClient); ok {
		if sc.opts.maxStreams != 5 {
			t.Errorf("Expected %d streams got %d", 5, sc.opts.maxStreams)
		}
	} else {
		t.Errorf("Expected *pdpStreamingClient from NewClient got %v", c)
	}
}
