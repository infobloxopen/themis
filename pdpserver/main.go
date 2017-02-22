package main

import log "github.com/Sirupsen/logrus"

func main() {
	InitLogging(config.Verbose)
	log.Info("Starting PDP server")

	pdp := NewServer(config.CWD)

	pdp.LoadPolicies(config.Policy)

	pdp.ListenRequests(config.ServiceEP)
	pdp.ListenControl(config.ControlEP)
	pdp.Serve()
}
