package client

// An Option allows to set PIP client options.
type Option func(*options)

// WithNetwork returns an Option which sets destination network.
func WithNetwork(n string) Option {
	return func(o *options) {
		o.net = n
	}
}

// WithAddress returns an Option which sets destination address.
func WithAddress(a string) Option {
	return func(o *options) {
		o.addr = a
	}
}

type options struct {
	net  string
	addr string
}

var defaults = options{
	net:  "tcp",
	addr: "localhost:5600",
}
