package main

import "flag"

const defaultInput = "errors.yaml"

type config struct {
	input string
}

var conf config

func init() {
	flag.StringVar(&conf.input, "i", defaultInput, "path to input file")
	flag.Parse()
}
