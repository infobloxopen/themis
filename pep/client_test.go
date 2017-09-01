package pep

import (
	pdp "github.com/infobloxopen/themis/pdp-service"
	"testing"
)

func TestTestClient(t *testing.T) {
	c := NewTestClient()

	c.NextResponse = &pdp.Response{Effect: pdp.Response_PERMIT,
		Obligation: []*pdp.Attribute{{"String", "string", "test"}}}

	res := &TestResponseStruct{}

	c.ModalValidate(nil, res)

	if !res.Effect {
		t.Errorf("Expected PERMIT but got %v", res.Effect)
	}

	if res.String != "test" {
		t.Errorf("Expected 'test' but got %q for res.String", res.String)
	}

	err := c.Connect()
	if err != nil {
		t.Errorf("Expected nil from Connect() but got %s", err)
	}

	c.Close()
}

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
