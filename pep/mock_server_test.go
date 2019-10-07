package pep

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	pbc "github.com/infobloxopen/themis/pdp-control"
	pbs "github.com/infobloxopen/themis/pdp-service"
)

// MockServer mocks a PDP server
type MockServer struct {
	t              *testing.T
	cancelableCtx  context.Context
	cancelCtxFn    context.CancelFunc
	listenAddrPort string
	grpcSvr        *grpc.Server
	listener       net.Listener
	validateSecs   int
}

// NewMockServer instantiates new instance of MockServer
// and spawns goroutine to serve incoming requests
func NewMockServer(listenAddrPort string, validateSecs int, t *testing.T) *MockServer {
	s := &MockServer{
                t:              t,
		listenAddrPort: listenAddrPort,
		validateSecs:   validateSecs,
	}

	cancelableCtx, cancelCtxFn := context.WithCancel(context.Background())
	s.cancelableCtx = cancelableCtx
	s.cancelCtxFn = cancelCtxFn

	go func() {
		if err := s.ServeRequests(); err != nil {
			t.Fatalf("ServeRequests failed: %s\n", err)
		}
	}()

	return s
}

// Stop stops the mock server
func (s *MockServer) Stop() {
	s.cancelCtxFn()
	if s.grpcSvr != nil {
		s.grpcSvr.Stop()
	}
}

// Request is GRPC handler for PDPControl service
func (s *MockServer) Request(ctx context.Context, in *pbc.Item) (*pbc.Response, error) {
	return &pbc.Response{}, nil
}

// Upload is GRPC handler for PDPControl service
func (s *MockServer) Upload(stream pbc.PDPControl_UploadServer) error {
	return nil
}

// Apply is GRPC handler for PDPControl service
func (s *MockServer) Apply(ctx context.Context, in *pbc.Update) (*pbc.Response, error) {
	return &pbc.Response{}, nil
}

// NotifyReady is GRPC handler for PDPControl service
func (s *MockServer) NotifyReady(ctx context.Context, m *pbc.Empty) (*pbc.Response, error) {
	return &pbc.Response{}, nil
}

// NewValidationStream is GRPC handler for PDP service
func (s *MockServer) NewValidationStream(stream pbs.PDP_NewValidationStreamServer) error {
	return nil
}

// Validate is GRPC handler for PDP service
func (s *MockServer) Validate(ctx context.Context, in *pbs.Msg) (*pbs.Msg, error) {
	timer := time.NewTimer(time.Duration(s.validateSecs) * time.Second)
	defer timer.Stop()

	select {
	case <-s.cancelableCtx.Done():
		s.t.Logf("Validate cancelled\n")
	case <-timer.C:
		s.t.Logf("Validate finished mock long-running processing\n")
	}

	return &pbs.Msg{}, nil
}

// ServeRequests serves PDP service requests
func (s *MockServer) ServeRequests() error {
	listener, err := net.Listen("tcp", "127.0.0.1:5555")
	if err != nil {
		s.t.Logf("net.Listen failed: %s\n", err)
		return err
	}

	s.grpcSvr = grpc.NewServer()
	s.listener = listener

	pbs.RegisterPDPServer(s.grpcSvr, s)
	defer s.grpcSvr.Stop()

	err = s.grpcSvr.Serve(s.listener)
	if err != nil {
		s.t.Logf("grpc.Serve failed: %s\n", err)
		return err
	}

	s.t.Logf("ServeRequests terminating\n")
	return nil
}
