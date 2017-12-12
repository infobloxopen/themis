package main

import (
	"flag"
	"log"
	"strings"
)

var (
	expr      string
	pattern   string
	rules     int
	count     int
	threshold int
	output    string

	policy   string
	rule     string
	value    string
	attrType string
)

func init() {
	flag.StringVar(&expr, "e", "regex", "use regex or wildcard expression")
	flag.StringVar(&pattern, "p", "prefix", "use prefix, infix or postfix pattern")
	flag.IntVar(&rules, "r", 1, "number of rules in policy")
	flag.IntVar(&count, "n", 1000, "number or requests to validate")
	flag.IntVar(&threshold, "t", 10, "number of requests to validate in parallel")
	flag.StringVar(&output, "o", "test.json", "file to dump validation timings")

	flag.Parse()

	e, ok := policies[strings.ToLower(expr)]
	if !ok {
		log.Fatalf("Unknown expression %q", expr)
	}

	p, ok := e[strings.ToLower(pattern)]
	if !ok {
		log.Fatalf("Unknown pattern %q", pattern)
	}

	policy = p[policyBody]
	rule = p[policyRule]
	value = p[attrValue]
	attrType = p[attrTypeIdx]
}
