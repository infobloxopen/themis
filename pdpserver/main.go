package main

import log "github.com/Sirupsen/logrus"

func main() {
	InitLogging(config.Verbose)
	log.Info("Starting PDP server")

	pdp := NewServer(config.CWD)

	pdp.LoadPolicies(config.Policy)

	pdp.ListenRequests(config.ServiceEP)
	pdp.ListenControl(config.ControlEP)

	tracer, err := InitTracing("zipkin", config.TracingEP)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Warning("Could not initialize tracing.")
	}
	pdp.Serve(tracer)
}
