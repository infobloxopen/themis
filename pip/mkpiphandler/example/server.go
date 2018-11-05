package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	"github.com/infobloxopen/themis/pip/mkpiphandler/example/pipexample"
	"github.com/infobloxopen/themis/pip/server"
)

func main() {
	log.Info("PIP example server")
	s := server.NewServer(
		server.WithConnErrHandler(errorLogger),
		server.WithHandler(pipexample.WrapHandler(handler)),
	)

	log.Info("Binding server")
	if err := s.Bind(); err != nil {
		log.WithError(err).Fatalf("failed to bind server")
	}

	log.Info("Serving requests")
	go func() {
		if err := s.Serve(); err != nil {
			log.WithError(err).Fatalf("failed to serve requests")
		}
	}()

	waitForInterrupt()

	log.Info("Stopping server")
	if err := s.Stop(); err != nil {
		log.WithError(err).Fatalf("failed to stop server")
	}

	log.Info("Done")
}

func handler(s string, addr net.IP) (*net.IPNet, error) {
	if s != serverKey {
		return nil, fmt.Errorf("unknown key %q", s)
	}

	if netIPv4.Contains(addr) {
		return netIPv4, nil
	}

	if netIPv6.Contains(addr) {
		return netIPv6, nil
	}

	return nil, fmt.Errorf("unknown address %q", addr)
}

const serverKey = "test"

var (
	netIPv4 *net.IPNet
	netIPv6 *net.IPNet
)

func init() {
	_, netIPv4, _ = net.ParseCIDR("192.0.2.0/24")
	_, netIPv6, _ = net.ParseCIDR("2001:db8::/32")
}

func errorLogger(addr net.Addr, err error) {
	if addr != nil {
		log.WithFields(log.Fields{
			"addr": addr,
			"err":  err,
		}).Error("server failure")
	} else {
		log.WithError(err).Error("server failure")
	}
}

func waitForInterrupt() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch
}
