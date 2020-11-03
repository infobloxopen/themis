package server

import (
	"strings"

	ot "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
)

func initTracing(tracingType, tracingEP string) (ot.Tracer, error) {
	if tracingEP == "" {
		return nil, nil
	}

	switch strings.ToLower(tracingType) {
	default:
		return nil, newTracingTypeError(tracingType)

	case "zipkin":
		return setupZipkin(tracingEP)
	}
}

func setupZipkin(tracingEP string) (ot.Tracer, error) {
	reporter := zipkinhttp.NewReporter(tracingEP)
	recorder, err := zipkin.NewEndpoint("PDP", "")
	if err != nil {
		return nil, err
	}
	tracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(recorder),
	)
	if err != nil {
		return nil, err
	}
	return zipkinot.Wrap(tracer), nil
}
