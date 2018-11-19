package pip

import (
	"testing"

	"github.com/infobloxopen/themis/pdp"
)

func TestClientsPoolGet(t *testing.T) {
	c0, err := pipClients.Get("localhost:5600")
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	}
	if c0 == nil {
		t.Error("expected instance of client.Client but got nothing")
	} else {
		defer c0.Close()

		c1, err := pipClients.Get("localhost:5600")
		if err != nil {
			t.Errorf("expected no error but got %#v", err)
		}
		if c0 != c1 {
			t.Errorf("expected the same client %#v but got %#v", c0, c1)
		}
	}
}

func TestClientsPoolRawGet(t *testing.T) {
	tc := new(testPipClient)

	pipClients.Lock()
	pipClients.m["localhost:5600"] = tc
	pipClients.Unlock()

	c, ok := pipClients.rawGet("localhost:5600")
	if !ok {
		t.Errorf("expected true but got %#v", ok)
	}
	if c != tc {
		t.Errorf("expected instance (%#v) of testPipClient but got %#v", tc, c)
	}

	c, ok = pipClients.rawGet("127.0.0.1:5600")
	if ok {
		t.Errorf("expected false but got %#v", ok)
	}
	if c != nil {
		t.Errorf("expected no testPipClient but got %#v", c)
	}
}

func TestClientsGetOrNew(t *testing.T) {
	tc := new(testPipClient)

	pipClients.Lock()
	pipClients.m["localhost:5600"] = tc
	pipClients.Unlock()

	c, err := pipClients.getOrNew("localhost:5600")
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	}
	if c != tc {
		t.Errorf("expected instance (%#v) of testPipClient but got %#v", tc, c)
	}

	c, err = pipClients.getOrNew("127.0.0.1:5600")
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	}
	if c == nil {
		t.Error("expected instance of client.Client but got nothing")
	} else {
		c.Close()
	}

	c, err = pipK8sClients.getOrNew("value.key.namespace:5600")
	if err == nil {
		t.Error("expected error")
	}
	if c != nil {
		t.Errorf("expected no client but got %#v", c)
		c.Close()
	}
}

type testPipClient struct{}

func (c *testPipClient) Connect() error { panic("not implemented") }
func (c *testPipClient) Close()         { panic("not implemented") }
func (c *testPipClient) Get(string, []pdp.AttributeValue) (pdp.AttributeValue, error) {
	panic("not implemented")
}
