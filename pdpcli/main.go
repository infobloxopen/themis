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
	client := NewClient()
	err := client.Connect(config.Server, config.Timeout)
	check(err, "can't connect to \"%s\"", config.Server)
	defer client.Close()

	reqs, err := LoadRequests(config.Input)
	check(err, "can't load requests from \"%s\"", config.Input)

	fmt.Printf("Got %d requests. Sending...\n", len(reqs.Requests))
	err = client.Send(reqs, config.Count, config.Output)
	check(err, "can't send requests")
}
