package main

import "flag"

type config struct {
	server string
	input  string
	count  int
	output string
}

var conf = config{}

func init() {
	flag.StringVar(&conf.server, "s", "127.0.0.1:5555", "PDP server to work with")
	flag.StringVar(&conf.input, "i", "requests.yaml", "file with YAML formatted list of requests to send to PDP")
	flag.IntVar(&conf.count, "n", 0, "number or requests to send "+
		"(default and value less than one means all requests from file)")
	flag.StringVar(&conf.output, "o", "", "file to write YAML formatted list of responses from PDP (default stdout)")
	flag.Parse()
}
