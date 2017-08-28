package main

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	ot "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	"github.com/infobloxopen/themis/pdp"
	pbc "github.com/infobloxopen/themis/pdp-control"
	pbs "github.com/infobloxopen/themis/pdp-service"
	"github.com/infobloxopen/themis/pdp/jcon"
	"github.com/infobloxopen/themis/pdp/yast"
)

type Transport struct {
	Interface net.Listener
	Protocol  *grpc.Server
}

type Server struct {
	sync.RWMutex

	Requests Transport
	Control  Transport
	Health   Transport

	queue *Queue

	p *pdp.PolicyStorage

	c  *pdp.LocalContentStorage
	ct map[string]*pdp.LocalContentStorageTransaction
}

func NewServer() *Server {
	return &Server{
		queue: NewQueue(),
		c:     pdp.NewLocalContentStorage(nil),
		ct:    make(map[string]*pdp.LocalContentStorageTransaction)}
}

func (s *Server) LoadPolicies(path string) error {
	if len(path) <= 0 {
		return nil
	}

	log.WithField("policy", path).Info("Loading policy")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.WithFields(log.Fields{"policy": path, "error": err}).Error("Failed load policy")
		return err
	}

	log.WithField("policy", path).Info("Parsing policy")
	p, err := yast.Unmarshal(b, nil)
	if err != nil {
		log.WithFields(log.Fields{"policy": path, "error": err}).Error("Failed parse policy")
		return err
	}

	s.p = p

	return nil
}

func (s *Server) LoadContent(paths []string) error {
	items := []*pdp.LocalContent{}
	for _, path := range paths {
		err := func() error {
			log.WithField("content", path).Info("Opening content")
			f, err := os.Open(path)
			if err != nil {
				return err
			}

			defer f.Close()

			log.WithField("content", path).Info("Parsing content")
			item, err := jcon.Unmarshal(f, nil)
			if err != nil {
				return err
			}

			items = append(items, item)
			return nil
		}()
		if err != nil {
			log.WithFields(log.Fields{"content": path, "error": err}).Error("Failed parse content")
			return err
		}
	}

	s.c = pdp.NewLocalContentStorage(items)

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
	if len(addr) <= 0 {
		return nil
	}

	log.WithField("address", addr).Info("Opening health check port")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open health check port")
		return err
	}

	s.Health.Interface = ln
	return nil
}

func (s *Server) Serve(tracer ot.Tracer, profiler string) {
	go func() {
		log.Info("Creating control protocol handler")
		s.Control.Protocol = grpc.NewServer()
		pbc.RegisterPDPControlServer(s.Control.Protocol, s)

		log.Info("Serving control requests")
		s.Control.Protocol.Serve(s.Control.Interface)
	}()

	if s.Health.Interface != nil {
		healthMux := http.NewServeMux()
		healthMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			log.Info("Health check responding with OK")
			io.WriteString(w, "OK")
		})
		go func() {
			http.Serve(s.Health.Interface, healthMux)
		}()
	}

	if len(profiler) > 0 {
		go func() {
			http.ListenAndServe(profiler, http.DefaultServeMux)
		}()
	}

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
