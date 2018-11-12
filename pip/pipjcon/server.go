package main

import (
	"net"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-control"
	"github.com/infobloxopen/themis/pdp/jcon"
	"github.com/infobloxopen/themis/pip/server"
)

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

type srv struct {
	sync.RWMutex

	ss *server.Server
	sc *grpc.Server

	c *pdp.LocalContentStorage
	a argsPool

	uIdx int32
	u    *update

	once *sync.Once
}

func newSrv() *srv {
	return &srv{
		c:    pdp.NewLocalContentStorage(nil),
		a:    makeArgsPool(conf.maxArgs),
		once: new(sync.Once),
	}
}

func (s *srv) load() {
	if len(conf.content) > 0 {
		log.WithField("content", conf.content).Info("opening content")
		f, err := os.Open(conf.content)
		if err != nil {
			log.WithError(err).Fatal("failed to open content")
		}
		defer f.Close()

		log.WithField("content", conf.content).Info("parsing content")
		item, err := jcon.Unmarshal(f, nil)
		if err != nil {
			log.WithError(err).Fatal("failed to parse content")
		}

		s.Lock()
		s.c = s.c.Add(item)
		s.Unlock()
	}
}

func (s *srv) start() {
	s.startCtrl()

	s.RLock()
	sc := s.sc
	s.RUnlock()

	if sc == nil || len(conf.content) > 0 {
		s.once.Do(s.startSrv)
	}
}

func (s *srv) stop() {
	s.Lock()
	ss := s.ss
	s.ss = nil
	sc := s.sc
	s.sc = nil
	s.Unlock()

	if ss != nil {
		if err := ss.Stop(); err != nil {
			log.WithError(err).Fatal("failed to stop service")
		}
	}

	if sc != nil {
		sc.Stop()
	}
}

func (s *srv) startSrv() {
	s.Lock()
	defer s.Unlock()

	s.ss = server.NewServer(
		server.WithNetwork(conf.net),
		server.WithAddress(conf.addr),
		server.WithMaxConnections(conf.maxConn),
		server.WithConnErrHandler(func(addr net.Addr, err error) {
			if addr != nil {
				log.WithFields(log.Fields{
					"addr": addr,
					"err":  err,
				}).Error("service failure")
			} else {
				log.WithError(err).Error("service failure")
			}
		}),
		server.WithBufferSize(conf.bufSize),
		server.WithMaxMessageSize(conf.maxMsgSize),
		server.WithWriteInterval(conf.writeInt),
		server.WithHandler(s.handler),
	)

	log.WithFields(log.Fields{
		"network": conf.net,
		"address": conf.addr,
	}).Info("opening service port")
	if err := s.ss.Bind(); err != nil {
		log.WithError(err).Fatal("failed to open service port")
	}

	go func(s *server.Server) {
		if err := s.Serve(); err != nil {
			log.WithError(err).Fatal("failed to start service")
		}
	}(s.ss)
}

func (s *srv) startCtrl() {
	s.Lock()
	defer s.Unlock()

	if ln := s.bindCtrl(); ln != nil {
		s.sc = grpc.NewServer()
		pb.RegisterPDPControlServer(s.sc, s)

		go func(s *grpc.Server) {
			if err := s.Serve(ln); err != nil {
				log.WithError(err).Fatal("failed to start control")
			}
		}(s.sc)
	}
}

func (s *srv) bindCtrl() net.Listener {
	if len(conf.ctrl) > 0 {
		log.WithField("address", conf.ctrl).Info("opening control port")
		ln, err := net.Listen(conf.net, conf.ctrl)
		if err != nil {
			log.WithError(err).Fatal("failed to open control port")
		}

		return ln
	}

	return nil
}
