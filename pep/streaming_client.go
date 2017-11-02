package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"errors"
	"fmt"
	"sync"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	ot "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-service"
)

type pdpStreamingClient struct {
	conn   *grpc.ClientConn
	client *pb.PDPClient

	opts options

	lock    *sync.RWMutex
	streams chan pb.PDP_NewValidationStreamClient
}

func (c *pdpStreamingClient) Connect(addr string) error {
	if c.conn != nil {
		return ErrorConnected
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	if len(c.opts.addresses) > 0 {
		addr = virtualServerAddress
		opts = append(opts, grpc.WithBalancer(grpc.RoundRobin(newStaticResolver(addr, c.opts.addresses...))))
	}

	if c.opts.tracer != nil {
		opts = append(opts,
			grpc.WithUnaryInterceptor(
				otgrpc.OpenTracingClientInterceptor(
					c.opts.tracer,
					otgrpc.IncludingSpans(
						func(parentSpanCtx ot.SpanContext, method string, req, resp interface{}) bool {
							return parentSpanCtx != nil
						},
					),
				),
			),
		)
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return err
	}

	c.conn = conn

	client := pb.NewPDPClient(c.conn)
	c.client = &client

	return nil
}

func (c *pdpStreamingClient) Close() {
	c.closeStreams()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.client = nil
}

func (c *pdpStreamingClient) Validate(in, out interface{}) error {
	s, err := c.getStream()
	if err != nil {
		return err
	}
	defer c.putStream(s)

	return s.validate(in, out)
}

func (c *pdpStreamingClient) makeStreams() error {
	if c.opts.maxStreams <= 0 {
		panic(fmt.Errorf("streaming client must be created with at least 1 stream but got %d", c.opts.maxStreams))
	}

	if c.client == nil {
		return ErrorNotConnected
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.streams != nil {
		return nil
	}

	c.streams = make(chan pb.PDP_NewValidationStreamClient, c.opts.maxStreams)
	for i := 0; i < c.opts.maxStreams; i++ {
		s, err := (*c.client).NewValidationStream(context.TODO(),
			grpc.FailFast(false),
		)
		if err != nil {
			return err
		}

		c.streams <- s
	}

	return nil
}

func (c *pdpStreamingClient) closeStreams() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.streams == nil {
		return
	}

	var wg sync.WaitGroup
	for len(c.streams) > 0 {
		s := <-c.streams
		if err := s.CloseSend(); err != nil {
			continue
		}

		wg.Add(1)
		go func(s pb.PDP_NewValidationStreamClient) {
			defer wg.Done()

			var msg interface{}
			s.RecvMsg(&msg)
		}(s)
	}
	wg.Wait()

	close(c.streams)
	c.streams = nil
}

func (c *pdpStreamingClient) getStream() (stream, error) {
	c.lock.RLock()
	ch := c.streams
	c.lock.RUnlock()

	if ch == nil {
		if err := c.makeStreams(); err != nil {
			c.closeStreams()
			return stream{}, err
		}

		c.lock.RLock()
		ch = c.streams
		c.lock.RUnlock()
	}

	if s, ok := <-ch; ok {
		return stream{s: s, ch: ch}, nil
	}

	return stream{}, ErrorNotConnected
}

func (c *pdpStreamingClient) putStream(s stream) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.streams == nil || c.streams != s.ch {
		return errors.New("invalid connection")
	}

	s.ch <- s.s
	return nil
}

type stream struct {
	s  pb.PDP_NewValidationStreamClient
	ch chan pb.PDP_NewValidationStreamClient
}

func (s stream) validate(in, out interface{}) error {
	req, err := makeRequest(in)
	if err != nil {
		return err
	}

	err = s.s.Send(&req)
	if err != nil {
		return err
	}

	res, err := s.s.Recv()
	if err != nil {
		return err
	}

	return fillResponse(res, out)
}
