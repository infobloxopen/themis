package main

import (
	"flag"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

type config struct {
	net        string
	addr       string
	ctrl       string
	content    string
	maxConn    int
	bufSize    int
	maxMsgSize int
	writeInt   time.Duration
}

const (
	netUnix  = "unix"
	netTCP   = "tcp"
	netTCPv4 = "tcp4"
	netTCPv6 = "tcp6"

	defUnixAddr = "/var/run/pip.socket"
	defTCPAddr  = "localhost:5600"
	defCtrlAddr = "localhost:5604"
)

var validNetworks = map[string]struct{}{
	netUnix:  {},
	netTCP:   {},
	netTCPv4: {},
	netTCPv6: {},
}

var conf config

func init() {
	flag.StringVar(&conf.net, "network", netTCP, "type of network to listen at")
	flag.StringVar(&conf.addr, "a", "", "address to listen at "+
		"(default for \"tcp*\" - localhost:5600, default for \"unix\" - /var/run/pip.socket)")
	flag.StringVar(&conf.ctrl, "c", "", "address for control "+
		"(default for \"tcp*\" - localhost:5604, unavailable for \"unix\")")
	flag.StringVar(&conf.content, "j", "", "path to JCon file to load at startup")
	flag.IntVar(&conf.maxConn, "max-connections", 0, "limit on number of simultaneous connections "+
		"(defailt - no limit)")
	flag.IntVar(&conf.bufSize, "buffer-size", 1024*1024, "input/output buffer size")
	flag.IntVar(&conf.maxMsgSize, "max-message", 10*1024, "limit on single request/response size")
	flag.DurationVar(&conf.writeInt, "write-interval", 50*time.Microsecond,
		"interval to wait for responses if output buffer isn't full")

	flag.Parse()

	netID := strings.ToLower(conf.net)
	if _, ok := validNetworks[netID]; !ok {
		log.WithField("network", conf.net).Fatal("unknown network type")
	}

	if len(conf.addr) <= 0 {
		if netID == netUnix {
			conf.addr = defUnixAddr
		} else {
			conf.addr = defTCPAddr
		}
	}

	if len(conf.ctrl) <= 0 {
		if netID != netUnix {
			conf.ctrl = defCtrlAddr
		}
	} else {
		if netID == netUnix {
			log.WithField("control", conf.ctrl).Info("control address set for \"unix\" network. ignoring...")
			conf.ctrl = ""
		}
	}
}
