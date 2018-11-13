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

func run() error {
	reqs, err := loadRequests(conf.Requests)
	if err != nil {
		return err
	}

	for i, r := range reqs {
		fmt.Printf("%d:\n\tpath: %q\n\targs:\n", i+1, r.path)
		for j, v := range r.args {
			s, err := v.Serialize()
			if err != nil {
				return fmt.Errorf("(%d:%d): %s", i+1, j+1, err)
			}
			fmt.Printf("\t\t%d: %q\n", j+1, s)
		}
		fmt.Println()
	}

	opts := []client.Option{
		client.WithNetwork(conf.Network),
		client.WithAddress(conf.Address),
		client.WithMaxRequestSize(conf.MaxRequestSize),
		client.WithMaxQueue(conf.MaxQueue),
		client.WithBufferSize(conf.BufferSize),
		client.WithConnErrHandler(connErrHandler),
		client.WithConnTimeout(conf.ConnTimeout),
		client.WithWriteInterval(conf.WriteInterval),
		client.WithResponseTimeout(conf.ResponseTimeout),
		client.WithResponseCheckInterval(conf.ResponseCheckInterval),
	}

	if conf.RoundRobinBalancer {
		opts = append(opts, client.WithRoundRobinBalancer(conf.Servers...))
	} else if conf.HotSpotBalancer {
		opts = append(opts, client.WithHotSpotBalancer(conf.Servers...))
	}

	if conf.DNSRadar {
		opts = append(opts, client.WithDNSRadar())
	} else if conf.K8sRadar {
		opts = append(opts, client.WithK8sRadar())
	}

	c := client.NewClient(opts...)
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
