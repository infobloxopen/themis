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

	pdp.listenRequests(conf.serviceEP)
	pdp.listenControl(conf.controlEP)
	pdp.listenHealthCheck(conf.healthEP)
	pdp.listenProfiler(conf.profilerEP)

	tracer, err := initTracing("zipkin", conf.tracingEP)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warning("Could not initialize tracing.")
	}
	pdp.serve(tracer)
}
