package pep

import "testing"

func TestNewBalancedClient(t *testing.T) {
	addrs := []string{"127.0.0.1:1000", "127.0.0.1:1001"}
	c := NewBalancedClient(addrs)
	cc, ok := c.(*client)
	if !ok {
		t.Errorf("Expected *client from NewBalancedClient() but got %v", c)
	} else {
		if cc.rpcs == nil {
			t.Errorf("Expected rps objects to be set but got nil")
		}
	}
}
