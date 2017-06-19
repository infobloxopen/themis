package main

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

import (
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"

	pbc "github.com/infobloxopen/themis/pdp-control"
	pbs "github.com/infobloxopen/themis/pdp-service"

	"github.com/infobloxopen/themis/pdp"

	"errors"
	ot "github.com/opentracing/opentracing-go"
	"io"
	"net/http"
)

type Transport struct {
	Interface net.Listener
	Protocol  *grpc.Server
}

type Server struct {
	sync.RWMutex

	Version    string
	Policy     pdp.EvaluableType
	Attributes map[string]pdp.AttributeType
	Includes   map[string]interface{}

	Requests Transport
	Control  Transport
	Health   Transport

	Updates *Queue

	AffectedPolicies map[string]pdp.ContentPolicyIndexItem

	Ctx pdp.YastCtx
}

func NewServer(path string) *Server {
	return &Server{
		Updates:          NewQueue(),
		AffectedPolicies: map[string]pdp.ContentPolicyIndexItem{},
		Ctx:              pdp.NewYASTCtx(path)}
}

func (s *Server) LoadPolicies(path string) error {
	if len(path) == 0 {
		log.Error("Invalid path specified. Failed to Load Policies.")
		return errors.New("Invalid path specified.")
	}

	log.WithField("policy", path).Info("Loading policy")
	p, err := s.Ctx.UnmarshalYASTFromFile(path)
	if err != nil {
		log.WithFields(log.Fields{"policy": path, "error": err}).Error("Failed load policy")
		return err
	}

	s.Policy = p

	return nil
}

func (s *Server) ListenRequests(addr string) error {
	log.WithField("address", addr).Info("Opening service port")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open service port")
		return err
	}

	s.Requests.Interface = ln
	return nil
}

func (s *Server) ListenControl(addr string) error {
	log.WithField("address", addr).Info("Opening control port")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open control port")
		return err
	}

	s.Control.Interface = ln
	return nil
}

func (s *Server) ListenHealthCheck(addr string) error {
	log.WithField("address", addr).Info("Opening health check port")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open health check port")
		return err
	}

	s.Health.Interface = ln
	return nil
}

func (s *Server) Serve(tracer ot.Tracer) {
	go func() {
		log.Info("Creating control protocol handler")
		s.Control.Protocol = grpc.NewServer()
		pbc.RegisterPDPControlServer(s.Control.Protocol, s)

		log.Info("Serving control requests")
		s.Control.Protocol.Serve(s.Control.Interface)
	}()

	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Info("Health check responding with 300")
		io.WriteString(w, "OK")
	})
	go func() {
		http.Serve(s.Health.Interface, healthMux)
	}()

	log.Info("Creating service protocol handler")
	if tracer == nil {
		s.Requests.Protocol = grpc.NewServer()
	} else {
		onlyIfParent := func(parentSpanCtx ot.SpanContext, method string, req, resp interface{}) bool {
			return parentSpanCtx != nil
		}
		intercept := otgrpc.OpenTracingServerInterceptor(tracer, otgrpc.IncludingSpans(onlyIfParent))
		s.Requests.Protocol = grpc.NewServer(grpc.UnaryInterceptor(intercept))
	}
	pbs.RegisterPDPServer(s.Requests.Protocol, s)

	log.Info("Serving decision requests")
	s.Requests.Protocol.Serve(s.Requests.Interface)
}
