package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/infobloxopen/themis/pip/mkpiphandler/pkg"
)

func main() {
	log.WithFields(log.Fields{
		"schema": conf.schema,
		"output": conf.dir,
	}).Info("making pip handler")

	s, err := pkg.NewSchemaFromFile(conf.schema)
	if err != nil {
		log.WithError(err).Fatal("failed to load schema")
	}

	if err = s.Generate(conf.dir); err != nil {
		log.WithFields(log.Fields{
			"schema": conf.schema,
			"err":    err,
		}).Fatal("failed to generate package")
	}
}
