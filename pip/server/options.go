package server

// Option configures how we set up PDP server.
type Option func(*options)

// WithNetwork returns an Option which sets service network. It supports "tcp", "tcp4", "tcp6" and "unix" netwroks.
func WithNetwork(net string) Option {
	return func(o *options) {
		o.net = net
	}
}

// WithAddress returns an Option which sets service endpoint.
func WithAddress(addr string) Option {
	return func(o *options) {
		o.addr = addr
	}
}

type options struct {
	net  string
	addr string
}

var defaults = options{
	net:  "tcp",
	addr: "localhost:5555",
}
