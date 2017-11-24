package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-service"
)

const connectionResetPercent float64 = 0.3

func makeStreamConns(addrs []string, streams int, tracer opentracing.Tracer) []*streamConn {
	total := len(addrs)
	if total > streams {
		total = streams
	}

	conns := make([]*streamConn, total)
	chunk := streams / total
	rem := streams % total
	for i := range conns {
		count := chunk
		if i < rem {
			count++
		}

		conns[i] = newStreamConn(addrs[i], count, tracer)
	}

	return conns
}

func startConnRetryWorkers(conns []*streamConn, crp *connRetryPool) {
	for _, c := range conns {
		limit := int(float64(len(c.streams))*connectionResetPercent + 0.5)
		if limit < 1 {
			limit = 1
		}

		go c.retryWorker(limit, crp)
	}
}

func closeStreamConns(conns []*streamConn) {
	for _, c := range conns {
		if c != nil {
			c.closeConn()
		}
	}
}

const (
	scisDisconnected uint32 = iota
	scisConnecting
	scisConnected
	scisClosing
	scisClosed
	scisBroken
)

var (
	errStreamConnWrongState = errors.New("can't make operation with the connection")
	errConnFailure          = errors.New("connection failed")
)

type streamConn struct {
	addr   string
	tracer opentracing.Tracer

	state *uint32

	conn    *grpc.ClientConn
	client  pb.PDPClient
	streams []*stream
	index   *atomic.Value
	retry   chan boundStream
}

func newStreamConn(addr string, streams int, tracer opentracing.Tracer) *streamConn {
	state := scisDisconnected

	c := &streamConn{
		addr:    addr,
		tracer:  tracer,
		state:   &state,
		streams: make([]*stream, streams),
		index:   new(atomic.Value),
		retry:   make(chan boundStream),
	}

	for i := range c.streams {
		c.streams[i] = c.newStream()
	}

	return c
}

func (c *streamConn) connect() error {
	if !atomic.CompareAndSwapUint32(c.state, scisDisconnected, scisConnecting) {
		return errStreamConnWrongState
	}

	for {
		if err := c.tryConnect(time.Second); err == nil {
			if atomic.CompareAndSwapUint32(c.state, scisConnecting, scisConnected) {
				return nil
			}

			c.closeConnInternal()
			return errStreamConnWrongState
		}

		if atomic.CompareAndSwapUint32(c.state, scisClosing, scisClosed) {
			return errStreamConnWrongState
		}
	}
}

func (c *streamConn) tryConnect(timeout time.Duration) error {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.FailOnNonTempDialError(true),
	}

	if c.tracer != nil {
		opts = append(opts,
			grpc.WithUnaryInterceptor(
				otgrpc.OpenTracingClientInterceptor(
					c.tracer,
					otgrpc.IncludingSpans(inclusionFunc),
				),
			),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.addr, opts...)
	if err != nil {
		return err
	}

	c.conn = conn
	c.client = pb.NewPDPClient(c.conn)

	err = c.connectStreams()
	if err != nil {
		c.closeStreams()
		c.conn.Close()
		c.conn = nil
		c.client = nil
		return err
	}

	index := make(chan int, len(c.streams))
	for i := range c.streams {
		index <- i
	}
	c.index.Store(index)

	return nil
}

func (c *streamConn) newValidationStream() (pb.PDP_NewValidationStreamClient, error) {
	state := atomic.LoadUint32(c.state)
	if state != scisConnected && state != scisConnecting {
		return nil, errStreamConnWrongState
	}

	return c.client.NewValidationStream(context.TODO())
}

func (c *streamConn) connectStreams() error {
	for _, s := range c.streams {
		err := s.connect()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *streamConn) withdrawStreams() {
	index := c.index.Load().(chan int)
	count := 0
	for len(index) > 0 {
		select {
		default:
		case <-index:
			count++
		}
	}

	if count < cap(index) {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()

			t := time.NewTimer(5 * time.Second)
			for i := 0; i < cap(index); i++ {
				select {
				case <-index:
				case <-t.C:
					return
				}
			}

			if !t.Stop() {
				<-t.C
			}
		}()
		wg.Wait()
	}

	close(index)
	index = nil
	c.index.Store(index)
}

func (c *streamConn) closeStreams() {
	wg := &sync.WaitGroup{}
	for _, s := range c.streams {
		wg.Add(1)
		go s.closeStream(wg)
	}

	wg.Wait()
}

func (c *streamConn) closeConn() {
	if atomic.CompareAndSwapUint32(c.state, scisConnecting, scisClosing) ||
		!atomic.CompareAndSwapUint32(c.state, scisConnected, scisClosing) {
		return
	}

	close(c.retry)
	c.closeConnInternal()
}

func (c *streamConn) closeConnInternal() {
	c.withdrawStreams()
	c.closeStreams()
	c.conn.Close()
	atomic.StoreUint32(c.state, scisClosed)
}

func (c *streamConn) markDisconnected() bool {
	if !atomic.CompareAndSwapUint32(c.state, scisConnected, scisBroken) {
		return false
	}

	index := c.index.Load().(chan int)
	close(index)
	index = nil
	c.index.Store(index)

	c.closeStreams()
	c.conn.Close()
	c.conn = nil
	c.client = nil

	atomic.StoreUint32(c.state, scisDisconnected)
	return true
}

func (c *streamConn) getStream() (boundStream, error) {
	index := c.index.Load().(chan int)
	if index != nil {
		if i, ok := <-index; ok {
			return boundStream{
				s:     c.streams[i],
				idx:   i,
				index: index,
			}, nil
		}
	}

	return boundStream{}, errStreamConnWrongState
}

func (c *streamConn) tryGetStream() (boundStream, bool, error) {
	index := c.index.Load().(chan int)
	if index != nil {
		select {
		default:
			return boundStream{}, false, nil

		case i, ok := <-index:
			if ok {
				return boundStream{
					s:     c.streams[i],
					idx:   i,
					index: index,
				}, true, nil
			}
		}
	}

	return boundStream{}, false, errStreamConnWrongState
}

func (c *streamConn) putStream(s boundStream) error {
	if atomic.LoadUint32(c.state) != scisConnected {
		return errStreamConnWrongState
	}

	current := c.index.Load().(chan int)
	if current != s.index {
		return errStreamConnWrongState
	}

	defer recover()
	s.index <- s.idx
	return nil
}

func (c *streamConn) validate(in, out interface{}) error {
	if atomic.LoadUint32(c.state) != scisConnected {
		return errStreamConnWrongState
	}

	s, err := c.getStream()
	if err != nil {
		return err
	}

	err = s.s.validate(in, out)
	if err != nil {
		if err == errStreamFailure {
			c.retry <- s
		}

		return err
	}

	c.putStream(s)
	return nil
}

func (c *streamConn) tryValidate(in, out interface{}) (bool, error) {
	if atomic.LoadUint32(c.state) != scisConnected {
		return false, errStreamConnWrongState
	}

	s, ok, err := c.tryGetStream()
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	err = s.s.validate(in, out)
	if err != nil {
		if err == errStreamFailure {
			c.retry <- s
		}

		return true, err
	}

	c.putStream(s)
	return true, nil
}

func (c *streamConn) retryWorker(limit int, crp *connRetryPool) {
	pool := newStreamRetryPool(limit)
	for s := range c.retry {
		if atomic.LoadUint32(c.state) != scisConnected {
			c.putStream(s)
			continue
		}

		ver, ok := pool.push(s.idx)
		if !ok {
			c.putStream(s)
			pool.flush()
			crp.put(c)
			continue
		}

		go func(s boundStream, ver uint64) {
			defer func() {
				c.putStream(s)
				pool.pop(s.idx, ver)
			}()

			s.s.drop()
			if err := s.s.connect(); err != nil {
				crp.put(c)
			}
		}(s, ver)
	}
}

func inclusionFunc(parentSpanCtx opentracing.SpanContext, method string, req, resp interface{}) bool {
	return parentSpanCtx != nil
}
