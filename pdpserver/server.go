package main

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

import (
	"io"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/debug"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	ot "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"

	"github.com/infobloxopen/themis/pdp"
	pbc "github.com/infobloxopen/themis/pdp-control"
	"github.com/infobloxopen/themis/pdp/jcon"
	"github.com/infobloxopen/themis/pdp/yast"
	"github.com/valyala/gorpc"
)

type transport struct {
	iface net.Listener
	proto *grpc.Server
}

type server struct {
	sync.RWMutex

	tracer ot.Tracer

	startOnce sync.Once

	control  transport
	health   transport
	profiler net.Listener

	q *queue

	p *pdp.PolicyStorage
	c *pdp.LocalContentStorage

	softMemWarn *time.Time
	backMemWarn *time.Time
	fragMemWarn *time.Time
	gcMax       int
	gcPercent   int
}

func newServer() *server {

	tracer, err := initTracing("zipkin", conf.tracingEP)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warning("Cannot initialize tracing.")
		tracer = nil
	}

	gcp := debug.SetGCPercent(-1)
	if gcp != -1 {
		debug.SetGCPercent(gcp)
	}
	if gcp > 50 {
		gcp = 50
	}

	return &server{
		tracer:    tracer,
		q:         newQueue(),
		c:         pdp.NewLocalContentStorage(nil),
		gcMax:     gcp,
		gcPercent: gcp}
}

func (s *server) loadPolicies(path string) error {
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

func (s *server) loadContent(paths []string) error {
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

func (s *server) listenControl(addr string) {
	log.WithField("address", addr).Info("Opening control port")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open control port")
	}

	s.control.iface = ln
}

func (s *server) listenHealthCheck(addr string) {
	if len(addr) <= 0 {
		return
	}

	log.WithField("address", addr).Info("Opening health check port")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open health check port")
	}

	s.health.iface = ln
}

func (s *server) listenProfiler(addr string) {
	if len(addr) <= 0 {
		return
	}

	log.WithField("address", addr).Info("Opening profiler port")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open profiler port")
	}

	s.profiler = ln
}

func (s *server) serveRequests() {
	addr := conf.serviceEP
	log.WithField("address", addr).Info("Opening service port")
	server := gorpc.NewTCPServer(addr, s.Validate)
	err := server.Serve()
	if err != nil {
		log.WithFields(log.Fields{"address": addr, "error": err}).Fatal("Failed to open service port")
	}
}

func (s *server) serve() {
	s.listenControl(conf.controlEP)
	s.listenHealthCheck(conf.healthEP)
	s.listenProfiler(conf.profilerEP)
	go func() {
		log.Info("Creating control protocol handler")
		s.control.proto = grpc.NewServer()
		pbc.RegisterPDPControlServer(s.control.proto, s)

		log.Info("Serving control requests")
		s.control.proto.Serve(s.control.iface)
	}()

	if s.health.iface != nil {
		healthMux := http.NewServeMux()
		healthMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			log.Info("Health check responding with OK")
			io.WriteString(w, "OK")
		})
		go func() {
			http.Serve(s.health.iface, healthMux)
		}()
	}

	if s.profiler != nil {
		go func() {
			http.Serve(s.profiler, nil)
		}()
	}

	if len(conf.policy) != 0 || len(conf.content) != 0 {
		// We already have policy info applied; supplied from local files,
		// pointed to by CLI options.
		s.startOnce.Do(s.serveRequests)
	} else {
		// serveRequests() will be executed by external request.
		log.Info("Waiting for policies to be applied.")
	}
}
