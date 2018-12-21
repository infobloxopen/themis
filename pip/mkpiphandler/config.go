package main

import "flag"

type config struct {
	schema string
	dir    string
}

var conf config

func init() {
	flag.StringVar(&conf.schema, "s", "schema.yaml", "schema of PIP handler to generate")
	flag.StringVar(&conf.dir, "d", ".", "directory to put generated PIP handler package")

	flag.Parse()
}
