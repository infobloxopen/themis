package pep

import "testing"

func TestNewClient(t *testing.T) {
	c := NewClient("127.0.0.1:1000", nil)
	cc, ok := c.(*pdpClient)
	if !ok {
		t.Errorf("Expected *pdpClient from NewClient got %v", c)
	} else {
		if cc.addr != "127.0.0.1:1000" {
			t.Errorf("Expected address of 127.0.0.1:1000 but got %s", cc.addr)
		}
	}
}

func TestNewBalancedClient(t *testing.T) {
	addrs := []string{"127.0.0.1:1000", "127.0.0.1:1001"}
	c := NewBalancedClient(addrs, nil)
	cc, ok := c.(*pdpClient)
	if !ok {
		t.Errorf("Expected *pdpClient from NewClient got %v", c)
	} else {
		if cc.balancer == nil {
			t.Errorf("Expected balancer to be set but got nil")
		}
	}
}
