package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"runtime"
)

func main() {
	initLogging(conf.verbose)
	log.Info("Starting PDP server")

	pdp := newServer()

	if pdp == nil {
		log.Error("Failed to create Server.")
		os.Exit(1)
	}

	pdp.loadPolicies(conf.policy)
	pdp.loadContent(conf.content)
	runtime.GC()

	if pdp.listenRequests(conf.serviceEP) != nil {
		log.Error("Failed to Listen to Requests.")
		os.Exit(1)
	}
	if pdp.listenControl(conf.controlEP) != nil {
		log.Error("Failed to Listen to Control Packets.")
		os.Exit(1)
	}
	if pdp.listenHealthCheck(conf.healthEP) != nil {
		log.Error("Failed to Listen to Health Check.")
		os.Exit(1)
	}

	tracer, err := initTracing("zipkin", conf.tracingEP)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warning("Could not initialize tracing.")
	}
	pdp.serve(tracer, conf.profilerEP)
}
