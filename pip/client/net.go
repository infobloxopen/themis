package client

import "net"

func lookupHostPort(addr string) ([]string, error) {
	h, p, err := splitHostPort(addr)
	if err != nil {
		return nil, err
	}

	addrs, err := net.LookupHost(h)
	if err != nil {
		return nil, err
	}

	return joinAddrsPort(addrs, p), nil
}

func splitHostPort(addr string) (string, string, error) {
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		if err, ok := err.(*net.AddrError); !ok || err.Err != "missing port in address" {
			return "", "", err
		}

		return addr, defPort, nil
	}

	return h, p, nil
}

func joinAddrsPort(addrs []string, port string) []string {
	out := make([]string, 0, len(addrs))
	for _, host := range addrs {
		if addr := joinAddrPort(host, port); len(addr) > 0 {
			out = append(out, addr)
		}
	}

	return out
}

func joinAddrPort(addr string, port string) string {
	if ip := net.ParseIP(addr); ip != nil {
		if ip.To4() != nil {
			return addr + ":" + port
		}

		return "[" + addr + "]:" + port
	}

	return ""
}
