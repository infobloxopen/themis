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

	pdp.serve()

	// Prevent the main Go-routine to be finished,
	// while other Go-routines are serving requests.
	select {}
}
