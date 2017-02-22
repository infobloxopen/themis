package main

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/policy-box/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/policy-box/proto/ $GOPATH/src/github.com/infobloxopen/policy-box/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/policy-box/pdp-service && ls $GOPATH/src/github.com/infobloxopen/policy-box/pdp-service"

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/policy-box/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/policy-box/proto/ $GOPATH/src/github.com/infobloxopen/policy-box/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/policy-box/pdp-control && ls $GOPATH/src/github.com/infobloxopen/policy-box/pdp-control"

import (
	"net"
	"os"

	log "github.com/Sirupsen/logrus"
	"google.golang.org/grpc"

	pbs "github.com/infobloxopen/policy-box/pdp-service"
	pbc "github.com/infobloxopen/policy-box/pdp-control"

	"github.com/infobloxopen/policy-box/pdp"
)

type Server struct {
	Path   string
	Policy pdp.EvaluableType

	Requests net.Listener
	Control  net.Listener

	RequestsProtocol *grpc.Server
	ControlProtocol  *grpc.Server
}

func NewServer(path string) *Server {
	return &Server{Path: path}
}

func (s *Server) LoadPolicies(path string) {
	if len(path) == 0 {
		return
	}

	log.WithField("policy", path).Info("Loading policy")
	p, err := pdp.UnmarshalYASTFromFile(path, s.Path)
	if err != nil {
		log.WithFields(log.Fields{"policy": path, "error": err}).Error("Failed load policy")
		return
	}

	s.Policy = p
}

func (s *Server) ListenRequests(addr string) {
	log.WithField("address", addr).Info("Opening service port")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open service port")
		os.Exit(1)
	}

	s.Requests = ln
}

func (s *Server) ListenControl(addr string) {
	log.WithField("address", addr).Info("Opening control port")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open control port")
		os.Exit(1)
	}

	s.Control = ln
}

func (s *Server) Serve() {
	go func () {
		log.Info("Creating control protocol handler")
		s.ControlProtocol = grpc.NewServer()
		pbc.RegisterPDPControlServer(s.ControlProtocol, s)

		log.Info("Serving control requests")
		s.ControlProtocol.Serve(s.Control)
	}()

	log.Info("Creating service protocol handler")
	s.RequestsProtocol = grpc.NewServer()
	pbs.RegisterPDPServer(s.RequestsProtocol, s)

	log.Info("Serving decision requests")
	s.RequestsProtocol.Serve(s.Requests)
}
