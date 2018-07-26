package server

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/themis/pip-service"
)

type transport struct {
	iface net.Listener
	proto *grpc.Server
}

type Option func(*options)

func WithLogger(logger *log.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithServiceAt(addr string) Option {
	return func(o *options) {
		o.service = addr
	}
}

type options struct {
	logger  *log.Logger
	service string
}

type Server struct {
	opts     options
	requests transport
}

func NewServer(opts ...Option) *Server {
	o := options{
		logger:  log.StandardLogger(),
		service: ":5555",
	}

	for _, opt := range opts {
		opt(&o)
	}

	return &Server{
		opts: o,
	}
}

func (s *Server) listenRequests() error {
	s.opts.logger.WithField("address", s.opts.service).Info("Opening service port")
	ln, err := net.Listen("tcp", s.opts.service)
	if err != nil {
		return err
	}

	s.requests.iface = ln
	return nil
}

func (s *Server) configureRequests() []grpc.ServerOption {
	return []grpc.ServerOption{}
}

func (s *Server) serveRequests() error {
	err := s.listenRequests()
	if err != nil {
		return err
	}

	s.opts.logger.Info("Serving decision requests")
	if err := s.requests.proto.Serve(s.requests.iface); err != nil {
		return err
	}

	return nil
}

func (s *Server) Serve() error {
	s.opts.logger.Info("Creating service protocol handler")
	s.requests.proto = grpc.NewServer(s.configureRequests()...)
	pb.RegisterPIPServer(s.requests.proto, s)
	defer s.requests.proto.Stop()

	return s.serveRequests()
}

func (s *Server) Stop() error {
	if s.requests.proto == nil {
		return fmt.Errorf("server hasn't been started")
	}

	s.requests.proto.Stop()
	return nil
}
