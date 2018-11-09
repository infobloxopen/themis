package main

import (
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pdp-control"
	"github.com/infobloxopen/themis/pip/server"
)

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-control && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/control.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-control && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-control"

type srv struct {
	sync.RWMutex

	ss *server.Server

	ln net.Listener
	sc *grpc.Server

	c *pdp.LocalContentStorage

	uIdx int32
	u    *update
}

func newSrv() *srv {
	return &srv{
		ss: server.NewServer(
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
			server.WithHandler(handler),
		),
		c: pdp.NewLocalContentStorage(nil),
	}
}

func (s *srv) start() {
	log.WithFields(log.Fields{
		"network": conf.net,
		"address": conf.addr,
	}).Info("opening service port")
	if err := s.ss.Bind(); err != nil {
		log.WithError(err).Fatal("failed to open service port")
	}

	if len(conf.ctrl) > 0 {
		log.WithField("address", conf.ctrl).Info("opening control port")
		ln, err := net.Listen(conf.net, conf.ctrl)
		if err != nil {
			log.WithError(err).Fatal("failed to open control port")
		}

		s.ln = ln
	}

	go func() {
		if err := s.ss.Serve(); err != nil {
			log.WithError(err).Fatal("failed to start service")
		}
	}()

	if s.ln != nil {
		s.sc = grpc.NewServer()
		pb.RegisterPDPControlServer(s.sc, s)

		go func() {
			if err := s.sc.Serve(s.ln); err != nil {
				log.WithError(err).Fatal("failed to start control")
			}
		}()
	}
}

func (s *srv) stop() {
	if err := s.ss.Stop(); err != nil {
		log.WithError(err).Fatal("failed to stop service")
	}

	if s.sc != nil {
		s.sc.Stop()
	}
}
