package main

import (
	"fmt"
	"net"
	"os"

	"github.com/infobloxopen/themis/pip/client"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var c client.Client

func run() error {
	if err := conf.LoadRequests(); err != nil {
		return err
	}

	c = conf.NewClient(connErrHandler)
	if err := c.Connect(); err != nil {
		return err
	}
	defer c.Close()

	return cmd(conf)
}

func connErrHandler(addr net.Addr, err error) {
	if addr != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", addr, err)
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}
