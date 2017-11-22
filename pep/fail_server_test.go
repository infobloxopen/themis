package pep

import (
	"errors"
	"io"
	"net"
	"strconv"
	"strings"
	"sync/atomic"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pdp-service"
)

const (
	failID = "fail"
	IDID   = "id"

	thisRequest = "this"
)

var requestedError = errors.New("failed as requested by client")

type failServer struct {
	ID       uint64
	failNext int32
	s        *grpc.Server
}

func newFailServer(addr string) (*failServer, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &failServer{s: grpc.NewServer()}
	pb.RegisterPDPServer(s.s, s)
	go s.s.Serve(ln)

	return s, nil
}

func (s *failServer) Stop() {
	s.s.Stop()
}

func (s *failServer) Validate(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	reqID := atomic.AddUint64(&s.ID, 1)

	targetID, fail := parseFailRequest(in)
	if fail == thisRequest && reqID == targetID {
		return nil, requestedError
	}

	return &pb.Response{
		Effect:     pb.Response_PERMIT,
		Reason:     "Ok",
		Obligation: in.Attributes,
	}, nil
}

func (s *failServer) NewValidationStream(stream pb.PDP_NewValidationStreamServer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		reqID := atomic.AddUint64(&s.ID, 1)
		targetID, fail := parseFailRequest(in)
		if fail == thisRequest && reqID == targetID {
			return requestedError
		}

		err = stream.Send(&pb.Response{
			Effect:     pb.Response_PERMIT,
			Reason:     "Ok",
			Obligation: in.Attributes,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func parseFailRequest(in *pb.Request) (uint64, string) {
	var targetID uint64
	fail := ""
	for _, attr := range in.Attributes {
		switch strings.ToLower(attr.Id) {
		case IDID:
			ID, err := strconv.ParseUint(attr.Value, 10, 64)
			if err == nil {
				targetID = ID
			}

		case failID:
			fail = strings.ToLower(attr.Value)
		}
	}

	return targetID, fail
}
