package main

import (
	"fmt"
	"strings"
	log "github.com/Sirupsen/logrus"

	ot "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

func InitTracing(tracingType, tracingEP string) (ot.Tracer, error) {
	if tracingEP == "" {
		return nil, nil
	}

	switch tracingType {
	case "zipkin":
		return setupZipkin(tracingEP)
	default:
		return nil, fmt.Errorf("Invalid tracing type: %s", tracingType)
	}
}

func setupZipkin(tracingEP string) (ot.Tracer, error) {
	if strings.Index(tracingEP, "http") == -1 {
		tracingEP = "http://" + tracingEP + "/api/v1/spans"
	}

	collector, err := zipkin.NewHTTPCollector(tracingEP)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to create zipkin HTTP Collector.")
		return nil, err
	}

	recorder := zipkin.NewRecorder(collector, false, "", "PDP")
	return zipkin.NewTracer(recorder)
}
