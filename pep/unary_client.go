package pep

import (
	"context"
	"fmt"
	"sync"

	"github.com/allegro/bigcache"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	ot "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-service"
)

type UnaryClient struct {
	lock   *sync.RWMutex
	conn   *grpc.ClientConn
	client *pb.PDPClient

	pool bytePool

	cache *bigcache.BigCache

	opts options
}

func newUnaryClient(opts options) *UnaryClient {
	c := &UnaryClient{
		lock: &sync.RWMutex{},
		opts: opts,
	}

	if !opts.autoRequestSize {
		c.pool = makeBytePool(int(opts.maxRequestSize), opts.noPool)
	}

	return c
}

func (c *UnaryClient) Connect(addr string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.conn != nil {
		return ErrorConnected
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	if len(c.opts.addresses) > 0 {
		addr = virtualServerAddress
		switch c.opts.balancer {
		default:
			panic(fmt.Errorf("invalid balancer %d", c.opts.balancer))

		case roundRobinBalancer:
			opts = append(opts, grpc.WithBalancer(grpc.RoundRobin(newStaticResolver(addr, c.opts.addresses...))))

		case hotSpotBalancer:
			return ErrorHotSpotBalancerUnsupported
		}
	}

	var interceptors []grpc.UnaryClientInterceptor

	if c.opts.tracer != nil {
		interceptors = append(interceptors,
			otgrpc.OpenTracingClientInterceptor(
				c.opts.tracer,
				otgrpc.IncludingSpans(
					func(parentSpanCtx ot.SpanContext, method string, req, resp interface{}) bool {
						return parentSpanCtx != nil
					},
				),
			))
	}

	if c.opts.clientUnaryInterceptors != nil {
		interceptors = append(interceptors, c.opts.clientUnaryInterceptors...)
	}
	opts = append(opts, grpc.WithUnaryInterceptor(
		grpc_middleware.ChainUnaryClient(interceptors...),
	))

	cache, err := newCacheFromOptions(c.opts)
	if err != nil {
		return err
	}

	ctx := c.opts.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return err
	}

	c.conn = conn
	c.cache = cache

	client := pb.NewPDPClient(c.conn)
	c.client = &client

	return nil
}

func (c *UnaryClient) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	if c.cache != nil {
		c.cache.Reset()
		c.cache = nil
	}

	c.client = nil
}

func (c *UnaryClient) ValidateContext(ctx context.Context, in, out interface{}) error {
	return c.validate(ctx, in, out)
}

// Validate is deprecated, use ValidateContext
func (c *UnaryClient) Validate(in, out interface{}) error {
	return c.validate(context.Background(), in, out)
}

func (c *UnaryClient) validate(ctx context.Context, in, out interface{}) error {
	c.lock.RLock()
	uc := c.client
	c.lock.RUnlock()

	if uc == nil {
		return ErrorNotConnected
	}

	var (
		req pb.Msg
		err error
	)

	if c.opts.autoRequestSize {
		req, err = makeRequest(in)
	} else {
		var b []byte
		switch in.(type) {
		default:
			b = c.pool.Get()
			defer c.pool.Put(b)

		case []byte, pb.Msg, *pb.Msg:
		}

		req, err = makeRequestWithBuffer(in, b)
	}
	if err != nil {
		return err
	}

	if c.cache != nil {
		var b []byte
		if b, err = c.cache.Get(string(req.Body)); err == nil {
			err = fillResponse(pb.Msg{Body: b}, out)
			if c.opts.onCacheHitHandler != nil {
				if err != nil {
					c.opts.onCacheHitHandler.Handle(in, b, err)
				} else {
					c.opts.onCacheHitHandler.Handle(in, out, nil)
				}
			}
			return err
		}
	}

	if c.opts.connTimeout > 0 {
		var cancelFn context.CancelFunc
		ctx, cancelFn = context.WithTimeout(ctx, c.opts.connTimeout)
		defer cancelFn()
	}

	res, err := (*uc).Validate(ctx, &req, grpc.FailFast(false))
	if err != nil {
		return err
	}

	if c.cache != nil {
		c.cache.Set(string(req.Body), res.Body)
	}

	return fillResponse(*res, out)
}
