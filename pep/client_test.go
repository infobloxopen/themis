package pep

import (
	"testing"

	ot "github.com/opentracing/opentracing-go"
)

func TestNewClient(t *testing.T) {
	c := NewClient()
	if _, ok := c.(*unaryClient); !ok {
		t.Errorf("Expected *unaryClient from NewClient got %#v", c)
	}
}

func TestNewBalancedClient(t *testing.T) {
	c := NewClient(WithRoundRobinBalancer("127.0.0.1:1000", "127.0.0.1:1001"))
	if uc, ok := c.(*unaryClient); ok {
		if len(uc.opts.addresses) <= 0 {
			t.Errorf("Expected balancer to be set but got nothing")
		}
	} else {
		t.Errorf("Expected *unaryClient from NewClient got %#v", c)
	}

	c = NewClient(WithHotSpotBalancer("127.0.0.1:1000", "127.0.0.1:1001"), WithStreams(5))
	if sc, ok := c.(*streamingClient); ok {
		if len(sc.opts.addresses) <= 0 {
			t.Errorf("Expected balancer to be set but got nothing")
		}
	} else {
		t.Errorf("Expected *streamingClient from NewClient got %#v", c)
	}
}

func TestNewStreamingClient(t *testing.T) {
	c := NewClient(WithStreams(5))
	if sc, ok := c.(*streamingClient); ok {
		if sc.opts.maxStreams != 5 {
			t.Errorf("Expected %d streams got %d", 5, sc.opts.maxStreams)
		}
	} else {
		t.Errorf("Expected *streamingClient from NewClient got %#v", c)
	}
}

func TestNewClientWithTracer(t *testing.T) {
	tr := &ot.NoopTracer{}
	c := NewClient(WithTracer(tr))
	uc, ok := c.(*unaryClient)
	if !ok {
		t.Fatalf("Expected *unaryClient from NewClient got %#v", c)
	}

	if uc.opts.tracer != tr {
		t.Errorf("Expected NoopTracer as client option but got %v", uc.opts.tracer)
	}
}

func TestNewClientWithMaxRequestSize(t *testing.T) {
	c := NewClient(WithMaxRequestSize(1024))
	uc, ok := c.(*unaryClient)
	if !ok {
		t.Fatalf("Expected *unaryClient from NewClient got %#v", c)
	}

	if uc.opts.maxRequestSize != 1024 {
		t.Errorf("Expected max size of %d bytes but got %d", 1024, uc.opts.maxRequestSize)
	}
}

func TestNewClientWithNoRequestBufferPool(t *testing.T) {
	c := NewClient(WithNoRequestBufferPool())
	uc, ok := c.(*unaryClient)
	if !ok {
		t.Fatalf("Expected *unaryClient from NewClient got %#v", c)
	}

	if uc.pool.b != nil {
		t.Errorf("Expected no pool but got %#v", uc.pool.b)
	}
}
