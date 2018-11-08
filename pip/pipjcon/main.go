package main

import log "github.com/sirupsen/logrus"

func main() {
	log.Info("PIP JCon server")

	s := newSrv()
	s.start()

	waitForInterrupt()

	s.stop()

	log.Info("done")
}
