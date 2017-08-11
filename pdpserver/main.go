package main

import (
	log "github.com/Sirupsen/logrus"
	"os"
	"runtime"
)

func main() {
	InitLogging(config.Verbose)
	log.Info("Starting PDP server")

	pdp := NewServer()

	if pdp == nil {
		log.Error("Failed to create Server.")
		os.Exit(1)
	}

	pdp.LoadPolicies(config.Policy)
	pdp.LoadContent(config.Content)
	runtime.GC()

	if pdp.ListenRequests(config.ServiceEP) != nil {
		log.Error("Failed to Listen to Requests.")
		os.Exit(1)
	}
	if pdp.ListenControl(config.ControlEP) != nil {
		log.Error("Failed to Listen to Control Packets.")
		os.Exit(1)
	}
	if pdp.ListenHealthCheck(config.HealthEP) != nil {
		log.Error("Failed to Listen to Health Check.")
		os.Exit(1)
	}

	tracer, err := InitTracing("zipkin", config.TracingEP)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warning("Could not initialize tracing.")
	}
	pdp.Serve(tracer, config.ProfilerEP)
}
