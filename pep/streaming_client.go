package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"errors"
	"fmt"
	"sync"
	"time"

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
	streams []pb.PDP_NewValidationStreamClient
	index   chan int
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
	ready := c.waitForStreams()

	c.lock.Lock()
	defer c.lock.Unlock()

	c.closeStreams(ready)

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

func (c *pdpStreamingClient) fillIndex() {
	c.index = make(chan int, len(c.streams))
	for i := range c.streams {
		c.index <- i
	}
}

func (c *pdpStreamingClient) dropIndex() {
	c.lock.Lock()
	defer c.lock.Unlock()

	close(c.index)
	c.index = nil
}

func (c *pdpStreamingClient) makeStreams() error {
	if c.opts.maxStreams <= 0 {
		panic(fmt.Errorf("streaming client must be created with at least 1 stream but got %d", c.opts.maxStreams))
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.client == nil {
		return ErrorNotConnected
	}

	if c.streams != nil {
		return nil
	}

	c.streams = make([]pb.PDP_NewValidationStreamClient, c.opts.maxStreams)
	ready := make([]int, len(c.streams))
	for i := range ready {
		ready[i] = -1
	}

	for i := range c.streams {
		s, err := (*c.client).NewValidationStream(context.TODO(),
			grpc.FailFast(false),
		)
		if err != nil {
			c.closeStreams(ready)
			return err
		}

		c.streams[i] = s
		ready[i] = i
	}

	c.fillIndex()

	return nil
}

func (c *pdpStreamingClient) waitForStreams() []int {
	c.lock.RLock()
	streams := c.streams
	index := c.index
	c.lock.RUnlock()

	if index == nil {
		return nil
	}

	ready := make([]int, len(streams))
	for i := range ready {
		ready[i] = -1
	}

	timeout := time.After(5 * time.Second)
	for i := range ready {
		select {
		case idx := <-index:
			ready[i] = idx

		case <-timeout:
			c.dropIndex()
			return ready
		}
	}

	c.dropIndex()
	return ready
}

func (c *pdpStreamingClient) closeStreams(ready []int) {
	if len(c.streams) <= 0 {
		return
	}

	wg := &sync.WaitGroup{}
	for _, i := range ready {
		if i >= 0 {
			wg.Add(1)
			go closeStream(c.streams[i], wg)
		}
	}
	wg.Wait()
	c.streams = nil
}

func closeStream(s pb.PDP_NewValidationStreamClient, wg *sync.WaitGroup) {
	defer wg.Done()

	if err := s.CloseSend(); err != nil {
		return
	}

	done := make(chan int)
	go func() {
		defer close(done)

		var msg pb.Response
		s.RecvMsg(&msg)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

func (c *pdpStreamingClient) getStream() (stream, error) {
	c.lock.RLock()
	streams := c.streams
	index := c.index
	c.lock.RUnlock()

	if streams == nil {
		if err := c.makeStreams(); err != nil {
			return stream{}, err
		}

		c.lock.RLock()
		streams = c.streams
		index = c.index
		c.lock.RUnlock()
	}

	if index != nil {
		if i, ok := <-index; ok {
			return stream{idx: i, s: streams[i], ch: index}, nil
		}
	}

	return stream{}, ErrorNotConnected
}

func (c *pdpStreamingClient) putStream(s stream) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.index == nil || c.index != s.ch {
		return errors.New("invalid connection")
	}

	s.ch <- s.idx
	return nil
}

type stream struct {
	idx int
	s   pb.PDP_NewValidationStreamClient
	ch  chan int
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
