package client

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

func lookupHostPort(addr string) ([]string, error) {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		if err, ok := err.(*net.AddrError); !ok || err.Err != "missing port in address" {
			return nil, err
		}

		h, p = addr, defPort
	}

	addrs, err := net.LookupHost(h)
	if err != nil {
		return nil, err
	}

	return joinAddrsPort(addrs, p), nil
}

func joinAddrsPort(addrs []string, port string) []string {
	out := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		if ip := net.ParseIP(addr); ip != nil {
			if ip.To4() != nil {
				out = append(out, addr+":"+port)
			} else {
				out = append(out, "["+addr+"]:"+port)
			}
		}
	}

	return out
}

func dialTimeout(network string, addresses []string, timeout time.Duration) ([]net.Conn, error) {
	conns := make([]net.Conn, len(addresses))
	errs := make([]error, len(addresses))

	wg := new(sync.WaitGroup)
	for i, address := range addresses {
		wg.Add(1)
		go func(i int, address string) {
			defer wg.Done()

			c, err := net.DialTimeout(network, address, timeout)
			if err != nil {
				errs[i] = err
			} else {
				conns[i] = c
			}
		}(i, address)
	}

	wg.Wait()

	n := 0
	for _, c := range conns {
		if c != nil {
			n++
		}
	}

	if n > 0 {
		out := make([]net.Conn, 0, n)
		for _, c := range conns {
			if c != nil {
				out = append(out, c)
			}
		}

		return out, nil
	}

	s := make([]string, len(errs))
	for i, err := range errs {
		s[i] = err.Error()
	}

	return nil, fmt.Errorf("failed to connect: \"%s\"", strings.Join(s, "\", \""))
}
