package client

import (
	"net"
	"time"
)

type dialer interface {
	dial(a string) (net.Conn, error)
}

type dialerTK struct {
	n string
	d *net.Dialer
}

func makeDialerTK(n string, t, k time.Duration) dialerTK {
	return dialerTK{
		n: n,
		d: &net.Dialer{
			Timeout:   t,
			KeepAlive: k,
		},
	}
}

func (d dialerTK) dial(a string) (net.Conn, error) {
	return d.d.Dial(d.n, a)
}
