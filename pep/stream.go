package pep

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/transport"

	pb "github.com/infobloxopen/themis/pdp-service"
)

const (
	ssDisconnected uint32 = iota
	ssConnecting
	ssConnected
	ssClosing
)

var (
	errStreamWrongState = errors.New("can't make operation with the stream")
	errStreamFailure    = errors.New("stream failed")
)

type stream struct {
	parent *streamConn
	state  *uint32

	stream pb.PDP_NewValidationStreamClient
}

func (c *streamConn) newStream() *stream {
	state := ssDisconnected
	return &stream{
		parent: c,
		state:  &state,
	}
}

func (s *stream) connect() error {
	if !atomic.CompareAndSwapUint32(s.state, ssDisconnected, ssConnecting) {
		return errStreamWrongState
	}

	var (
		ss  pb.PDP_NewValidationStreamClient
		err error
	)

	exitState := ssDisconnected
	defer func() {
		if atomic.CompareAndSwapUint32(s.state, ssClosing, ssDisconnected) {
			s.stream = nil
			if ss == nil {
				return
			}

			go func() {
				if err := ss.CloseSend(); err != nil {
					return
				}

				ss.Recv()
			}()

			return
		}

		atomic.StoreUint32(s.state, exitState)
	}()

	ss, err = s.parent.newValidationStream()
	if err != nil {
		return err
	}

	s.stream = ss
	exitState = ssConnected
	return nil
}

func (s *stream) closeStream(wg *sync.WaitGroup) {
	defer wg.Done()

	if atomic.CompareAndSwapUint32(s.state, ssConnecting, ssClosing) ||
		!atomic.CompareAndSwapUint32(s.state, ssConnected, ssClosing) {
		return
	}

	if err := s.stream.CloseSend(); err != nil {
		return
	}

	done := make(chan int)
	go func(s pb.PDP_NewValidationStreamClient) {
		defer close(done)
		s.Recv()
	}(s.stream)

	s.stream = nil

	t := time.NewTimer(5 * time.Second)
	select {
	case <-done:
		if !t.Stop() {
			<-t.C
		}
	case <-t.C:
	}

	atomic.StoreUint32(s.state, ssDisconnected)
}

func (s *stream) drop() {
	if atomic.CompareAndSwapUint32(s.state, ssConnecting, ssClosing) ||
		!atomic.CompareAndSwapUint32(s.state, ssConnected, ssClosing) {
		return
	}

	s.stream = nil
	atomic.StoreUint32(s.state, ssDisconnected)
}

func (s *stream) validate(in, out interface{}) error {
	if atomic.LoadUint32(s.state) != ssConnected {
		return errStreamWrongState
	}

	stream := s.stream

	req, err := makeRequest(in)
	if err != nil {
		return err
	}

	err = stream.Send(&req)
	if err != nil {
		if err == transport.ErrConnClosing || err == balancer.ErrTransientFailure {
			return errConnFailure
		}

		return errStreamFailure
	}

	res, err := stream.Recv()
	if err != nil {
		if err == transport.ErrConnClosing || err == balancer.ErrTransientFailure {
			return errConnFailure
		}

		return errStreamFailure
	}

	return fillResponse(res, out)
}

type boundStream struct {
	s     *stream
	idx   int
	index chan int
}
