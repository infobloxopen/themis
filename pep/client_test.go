package pep

import "testing"

func TestNewClient(t *testing.T) {
	c := NewClient()
	_, ok := c.(*pdpClient)
	if !ok {
		t.Errorf("Expected *pdpClient from NewClient got %v", c)
	}
}

func TestNewBalancedClient(t *testing.T) {
	c := NewClient(WithBalancer("127.0.0.1:1000", "127.0.0.1:1001"))
	cc, ok := c.(*pdpClient)
	if !ok {
		t.Errorf("Expected *pdpClient from NewClient got %v", c)
	} else {
		if cc.opts.balancer == nil {
			t.Errorf("Expected balancer to be set but got nil")
		}
	}
}
