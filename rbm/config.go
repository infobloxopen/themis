package main

import "flag"

var (
	count     int
	threshold int
	output    string
)

func init() {
	flag.IntVar(&count, "n", 1000, "number or requests to validate")
	flag.IntVar(&threshold, "t", 10, "number of requests to validate in parallel")
	flag.StringVar(&output, "o", "test.json", "file to dump validation timings")

	flag.Parse()
}
