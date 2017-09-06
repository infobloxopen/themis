package main

import (
	"fmt"
	"os"
)

func check(err error, format string, args ...interface{}) {
	if err == nil {
		return
	}

	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err)
	os.Exit(1)
}

func main() {
	reqs, err := loadRequests(conf.input)
	check(err, "can't load requests from \"%s\"", conf.input)

	fmt.Printf("Got %d requests. Sending...\n", len(reqs.Requests))
	err = send(conf.server, reqs, conf.count, conf.output)
	check(err, "can't send requests")
}
