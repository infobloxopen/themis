package main

import (
	"runtime"

	log "github.com/Sirupsen/logrus"
)

func main() {
	logger := log.StandardLogger()
	logger.Info("Starting PDP server.")

	pdp := NewServer(
		WithLogger(logger),
		WithPolicyParser(conf.policyParser),
		WithServiceAt(conf.serviceEP),
		WithControlAt(conf.controlEP),
		WithHealthAt(conf.healthEP),
		WithProfilerAt(conf.profilerEP),
		WithTracingAt(conf.tracingEP),
		WithMemLimits(conf.mem),
		WithMaxGRPCStreams(uint32(conf.maxStreams)),
	)

	err := pdp.LoadPolicies(conf.policy)
	if err != nil {
		logger.WithFields(
			log.Fields{
				"policy": conf.policy,
				"err":    err,
			},
		).Error("Failed to load policy. Continue with no policy...")
	}

	err = pdp.LoadContent(conf.content)
	if err != nil {
		logger.WithField("err", err).Error("Failed to load content. Continue with no content...")
	}

	runtime.GC()

	err = pdp.Serve()
	if err != nil {
		logger.WithError(err).Error("Failed to run server")
	}
}
