package main

import (
	"flag"
	"strings"
)

type Config struct {
	Verbose    int
	Policy     string
	Content    StringSet
	ServiceEP  string
	ControlEP  string
	TracingEP  string
	HealthEP   string
	ProfilerEP string
}

type StringSet []string

func (s *StringSet) String() string {
	return strings.Join(*s, ", ")
}

func (s *StringSet) Set(v string) error {
	*s = append(*s, v)
	return nil
}

var config Config

func init() {
	flag.IntVar(&config.Verbose, "v", 1, "log verbosity (0 - error, 1 - warn (default), 2 - info, 3 - debug)")
	flag.StringVar(&config.Policy, "p", "", "policy file to start with")
	flag.Var(&config.Content, "j", "JSON content files to start with")
	flag.StringVar(&config.ServiceEP, "l", "0.0.0.0:5555", "listen for decision requests on this address:port")
	flag.StringVar(&config.ControlEP, "c", "0.0.0.0:5554", "listen for policies on this address:port")
	flag.StringVar(&config.TracingEP, "t", "", "OpenZipkin tracing endpoint")
	flag.StringVar(&config.HealthEP, "health", "", "Health check endpoint")
	flag.StringVar(&config.ProfilerEP, "pprof", "", "Performance profiler endpoint")

	flag.Parse()
}
