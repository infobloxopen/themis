package main

import (
	"flag"
	"time"
)

type Config struct {
	Server  string
	Timeout time.Duration
	Input   string
	Count   int
	Output  string
}

var config = Config{}

func init() {
	flag.StringVar(&config.Server, "s", "127.0.0.1:5555", "PDP server to work with")
	flag.DurationVar(&config.Timeout, "timeout", 10*time.Second, "connection timeout")
	flag.StringVar(&config.Input, "i", "requests.yaml", "file with YAML formatted list of requests to send to PDP")
	flag.IntVar(&config.Count, "n", 0, "number or requests to send "+
		"(default and value less than one means all requests from file)")
	flag.StringVar(&config.Output, "o", "", "file to write YAML formatted list of responses from PDP (default stdout)")
	flag.Parse()
}
