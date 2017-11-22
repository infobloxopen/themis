package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	ot "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-service"
)

type validator func(in, out interface{}, clients []*subClient, counter *uint64) error

type streamingClient struct {
	opts options

	lock    *sync.RWMutex
	clients []*subClient
	counter uint64

	validate validator
}

func newStreamingClient(opts options) *streamingClient {
	if opts.maxStreams <= 0 {
		panic(fmt.Errorf("streaming client must be created with at least 1 stream but got %d", opts.maxStreams))
	}

	return &streamingClient{
		opts: opts,
		lock: &sync.RWMutex{},
	}
}

func (c *streamingClient) Connect(addr string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.clients) > 0 {
		return ErrorConnected
	}

	addrs := c.opts.addresses
	c.validate = simpleValidator
	if len(addrs) > 1 {
		switch c.opts.balancer {
		default:
			panic(fmt.Errorf("Invalid balancer %d", c.opts.balancer))

		case roundRobinBalancer:
			c.validate = roundRobinValidator

		case hotSpotBalancer:
			c.validate = hotSpotValidator
		}
	} else if len(addrs) <= 0 {
		addrs = []string{addr}
	}

	clients, err := makeClients(addrs, c.opts.maxStreams, c.opts.tracer)
	if err != nil {
		return err
	}

	c.clients = clients
	return nil
}

func (c *streamingClient) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	closeClients(c.clients)
	c.clients = nil
}

func (c *streamingClient) Validate(in, out interface{}) error {
	c.lock.RLock()
	clients := c.clients
	c.lock.RUnlock()

	if len(clients) <= 0 {
		return ErrorNotConnected
	}

	for {
		err := c.validate(in, out, clients, &c.counter)
		if err == nil {
			break
		}

		if err != streamError {
			return err
		}
	}

	return nil
}

func simpleValidator(in, out interface{}, clients []*subClient, counter *uint64) error {
	return clients[0].validate(in, out)
}

func roundRobinValidator(in, out interface{}, clients []*subClient, counter *uint64) error {
	i := int((atomic.AddUint64(counter, 1) - 1) % uint64(len(clients)))
	return clients[i].validate(in, out)
}

func hotSpotValidator(in, out interface{}, clients []*subClient, counter *uint64) error {
	i := 0

	total := uint64(len(clients))
	i = int(atomic.LoadUint64(counter) % total)
	j := i
	for {
		ok, err := clients[j].tryValidate(in, out)
		if ok {
			return err
		}

		j = int(atomic.AddUint64(counter, 1) % total)

		next := j + 1
		if next >= len(clients) {
			next = 0
		}

		if next == i {
			i = j
			break
		}
	}

	return clients[i].validate(in, out)
}

func makeClients(addrs []string, maxStreams int, tracer ot.Tracer) ([]*subClient, error) {
	total := len(addrs)
	if total > maxStreams {
		total = maxStreams
	}

	out := make([]*subClient, total)
	chunk := maxStreams / total
	rem := maxStreams % total
	for i := range out {
		count := chunk
		if i < rem {
			count++
		}

		sc, err := newSubClient(addrs[i], count, tracer)
		if err != nil {
			closeClients(out)
			return nil, err
		}

		out[i] = sc
	}

	return out, nil
}

func closeClients(clients []*subClient) {
	for _, sc := range clients {
		if sc != nil {
			sc.closeConn()
		}
	}
}

type subClient struct {
	conn   *grpc.ClientConn
	client *pb.PDPClient

	lock *sync.RWMutex

	maxStreams int
	streams    []pb.PDP_NewValidationStreamClient
	index      chan int

	retry chan stream
}

func newSubClient(addr string, maxStreams int, tracer ot.Tracer) (*subClient, error) {
	c := &subClient{
		lock:       &sync.RWMutex{},
		maxStreams: maxStreams,
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	if tracer != nil {
		opts = append(opts,
			grpc.WithUnaryInterceptor(
				otgrpc.OpenTracingClientInterceptor(
					tracer,
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
		return nil, err
	}

	c.conn = conn

	client := pb.NewPDPClient(c.conn)
	c.client = &client

	return c, nil
}

func (c *subClient) closeConn() {
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

func (c *subClient) validate(in, out interface{}) error {
	s, err := c.getStream()
	if err != nil {
		return err
	}

	err = s.validate(in, out)
	if err == streamError {
		c.retryStream(s)
		return err
	}

	c.putStream(s)
	return nil
}

func (c *subClient) tryValidate(in, out interface{}) (bool, error) {
	s, ok, err := c.tryGetStream()
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	err = s.validate(in, out)
	if err == streamError {
		c.retryStream(s)
		return true, err
	}

	c.putStream(s)
	return true, nil
}

func (c *subClient) fillIndex() {
	c.index = make(chan int, len(c.streams))
	for i := range c.streams {
		c.index <- i
	}
}

func (c *subClient) startRetryHandler() {
	c.retry = make(chan stream, len(c.streams))
	go func(ch chan stream) {
		for s := range ch {
			ss, err := (*s.client).NewValidationStream(context.TODO(),
				grpc.FailFast(false),
			)
			if err != nil {
				continue
			}

			s.s = ss
			c.putRestoredStream(s)
		}
	}(c.retry)
}

func (c *subClient) putRestoredStream(s stream) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.streams == nil || c.index == nil || c.index != s.ch {
		return errors.New("invalid connection")
	}

	s.streams[s.idx] = s.s
	s.ch <- s.idx
	return nil
}

func (c *subClient) dropIndex() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.index != nil {
		close(c.index)
		c.index = nil
	}
}

func (c *subClient) dropRetry() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.retry != nil {
		close(c.retry)
		c.retry = nil
	}
}

func (c *subClient) makeStreams() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.client == nil {
		return ErrorNotConnected
	}

	if c.streams != nil {
		return nil
	}

	c.streams = make([]pb.PDP_NewValidationStreamClient, c.maxStreams)
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
	c.startRetryHandler()

	return nil
}

func (c *subClient) waitForStreams() []int {
	c.dropRetry()

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

func (c *subClient) closeStreams(ready []int) {
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

func (c *subClient) getStream() (stream, error) {
	c.lock.RLock()
	streams := c.streams
	index := c.index
	retry := c.retry
	c.lock.RUnlock()

	if streams == nil {
		if err := c.makeStreams(); err != nil {
			return stream{}, err
		}

		c.lock.RLock()
		streams = c.streams
		index = c.index
		retry = c.retry
		c.lock.RUnlock()
	}

	if index != nil {
		if i, ok := <-index; ok {
			return stream{
				client:  c.client,
				streams: streams,
				idx:     i,
				s:       streams[i],
				ch:      index,
				retry:   retry,
			}, nil
		}
	}

	return stream{}, ErrorNotConnected
}

func (c *subClient) tryGetStream() (stream, bool, error) {
	c.lock.RLock()
	streams := c.streams
	index := c.index
	retry := c.retry
	c.lock.RUnlock()

	if streams == nil {
		if err := c.makeStreams(); err != nil {
			return stream{}, false, err
		}

		c.lock.RLock()
		streams = c.streams
		index = c.index
		retry = c.retry
		c.lock.RUnlock()
	}

	if index == nil {
		return stream{}, false, ErrorNotConnected
	}

	select {
	default:
		return stream{}, false, nil

	case i, ok := <-index:
		if ok {
			return stream{
				client:  c.client,
				streams: streams,
				idx:     i,
				s:       streams[i],
				ch:      index,
				retry:   retry,
			}, true, nil
		}
	}

	return stream{}, false, ErrorNotConnected
}

func (c *subClient) putStream(s stream) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.index == nil || c.index != s.ch {
		return errors.New("invalid connection")
	}

	s.ch <- s.idx
	return nil
}

func (c *subClient) retryStream(s stream) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	if c.retry == nil || c.retry != s.retry {
		return errors.New("invalid connection")
	}

	s.retry <- s
	return nil
}

type stream struct {
	client  *pb.PDPClient
	streams []pb.PDP_NewValidationStreamClient
	idx     int
	s       pb.PDP_NewValidationStreamClient
	ch      chan int

	retry chan stream
}

var streamError = errors.New("sending or receiving error for a stream")

func (s stream) validate(in, out interface{}) error {
	req, err := makeRequest(in)
	if err != nil {
		return err
	}

	err = s.s.Send(&req)
	if err != nil {
		return streamError
	}

	res, err := s.s.Recv()
	if err != nil {
		return streamError
	}

	return fillResponse(res, out)
}
