package main

import (
	"flag"
	"strings"
)

type config struct {
	verbose    int
	policy     string
	content    stringSet
	serviceEP  string
	controlEP  string
	tracingEP  string
	healthEP   string
	profilerEP string
	mem        memCheckConfig
}

type stringSet []string

func (s *stringSet) String() string {
	return strings.Join(*s, ", ")
}

func (s *stringSet) Set(v string) error {
	*s = append(*s, v)
	return nil
}

var conf config

func init() {
	flag.IntVar(&conf.verbose, "v", 1, "log verbosity (0 - error, 1 - warn (default), 2 - info, 3 - debug)")
	flag.StringVar(&conf.policy, "p", "", "policy file to start with")
	flag.Var(&conf.content, "j", "JSON content files to start with")
	flag.StringVar(&conf.serviceEP, "l", "0.0.0.0:5555", "listen for decision requests on this address:port")
	flag.StringVar(&conf.controlEP, "c", "0.0.0.0:5554", "listen for policies on this address:port")
	flag.StringVar(&conf.tracingEP, "t", "", "OpenZipkin tracing endpoint")
	flag.StringVar(&conf.healthEP, "health", "", "Health check endpoint")
	flag.StringVar(&conf.profilerEP, "pprof", "", "Performance profiler endpoint")
	flag.Uint64Var(&conf.mem.limit, "mem-limit", 0, "Memory limit in megabytes")

	flag.Parse()

	conf.mem.limit *= 1024 * 1024
	conf.mem.reset = 0.9 * float64(conf.mem.limit)
	conf.mem.soft = 0.7 * float64(conf.mem.limit)
	conf.mem.frag = 0.3
	conf.mem.back = 0.3
}
