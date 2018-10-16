package client

import "net"

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
