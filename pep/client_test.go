package pep

import (
	"context"
	"testing"
	"time"

	ot "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

func TestNewClient(t *testing.T) {
	c := NewClient()
	if _, ok := c.(*UnaryClient); !ok {
		t.Errorf("Expected *UnaryClient from NewClient got %#v", c)
	}
}

func TestNewBalancedClient(t *testing.T) {
	c := NewClient(WithRoundRobinBalancer("127.0.0.1:1000", "127.0.0.1:1001"))
	if uc, ok := c.(*UnaryClient); ok {
		if len(uc.opts.addresses) <= 0 {
			t.Errorf("Expected balancer to be set but got nothing")
		}
	} else {
		t.Errorf("Expected *UnaryClient from NewClient got %#v", c)
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
	uc, ok := c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}

	if uc.opts.tracer != tr {
		t.Errorf("Expected NoopTracer as client option but got %v", uc.opts.tracer)
	}
}

var noOpClientInterceptor grpc.UnaryClientInterceptor = func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return nil
}

func TestNewClientWithInterceptor(t *testing.T) {
	ci := noOpClientInterceptor
	c := NewClient(WithClientUnaryInterceptors(ci))
	uc, ok := c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}

	if len(uc.opts.clientUnaryInterceptors) == 0 {
		t.Errorf("Expected noOpClientInterceptor as client option but got %v", uc.opts.clientUnaryInterceptors)
	}
}

func TestNewClientWithAutoRequestSize(t *testing.T) {
	c := NewClient(WithAutoRequestSize(true))
	uc, ok := c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}

	if !uc.opts.autoRequestSize {
		t.Error("Expected auto size to be set")
	}
}

func TestNewClientWithMaxRequestSize(t *testing.T) {
	c := NewClient(WithMaxRequestSize(1024))
	uc, ok := c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}

	if uc.opts.maxRequestSize != 1024 {
		t.Errorf("Expected max size of %d bytes but got %d", 1024, uc.opts.maxRequestSize)
	}
}

func TestNewClientWithNoRequestBufferPool(t *testing.T) {
	c := NewClient(WithNoRequestBufferPool())
	uc, ok := c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}

	if uc.pool.b != nil {
		t.Errorf("Expected no pool but got %#v", uc.pool.b)
	}
}

func TestNewClientWithCacheTTL(t *testing.T) {
	c := NewClient(WithCacheTTL(5 * time.Second))
	uc, ok := c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}

	if !uc.opts.cache || uc.opts.cacheTTL != 5*time.Second {
		t.Errorf("Expected cache with TTL %s but got %#v, %s", 5*time.Second, uc.opts.cache, uc.opts.cacheTTL)
	}
}

func TestNewClientWithCacheTTLAndMaxSize(t *testing.T) {
	c := NewClient(WithCacheTTLAndMaxSize(5*time.Second, 1024))
	uc, ok := c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}

	if !uc.opts.cache || uc.opts.cacheTTL != 5*time.Second || uc.opts.cacheMaxSize != 1024 {
		t.Errorf("Expected cache with TTL %s and size limit %d but got %#v, %s, %d",
			5*time.Second, 1024, uc.opts.cache, uc.opts.cacheTTL, uc.opts.cacheMaxSize)
	}
}

func TestNewClientWithContext(t *testing.T) {
	c := NewClient()
	uc, ok := c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}
	if uc.opts.ctx != nil {
		t.Errorf("Expected default client to have nil context")
	}

	c = NewClient(WithContext(nil))
	uc, ok = c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}
	if uc.opts.ctx != nil {
		t.Errorf("Expected nil context to default to nil context")
	}

	toCtx, toCancelFn := context.WithTimeout(context.Background(), 1*time.Second)
	defer toCancelFn()
	c = NewClient(WithContext(toCtx))
	uc, ok = c.(*UnaryClient)
	if !ok {
		t.Fatalf("Expected *UnaryClient from NewClient got %#v", c)
	}
	if uc.opts.ctx != toCtx {
		t.Errorf("Expected timeout context")
	}
}
