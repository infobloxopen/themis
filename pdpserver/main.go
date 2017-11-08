package main

import (
	"runtime"

	log "github.com/Sirupsen/logrus"
)

func main() {
	initLogging(conf.verbose)
	log.Info("Starting PDP server.")

	pdp := newServer()

	if pdp == nil {
		log.Fatal("Failed to create PDP server.")
	}

	if err := pdp.setPolicyFormat(conf.policyFmt); err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Fatal("Failed to initialize PDP server.")
	}

	pdp.loadPolicies(conf.policy)
	pdp.loadContent(conf.content)
	runtime.GC()

	pdp.serve()
}
