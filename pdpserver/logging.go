package main

import log "github.com/Sirupsen/logrus"

func InitLogging(verbose bool) {
	level := log.WarnLevel
	if verbose {
		level = log.InfoLevel
	}

	log.SetLevel(level)
}
