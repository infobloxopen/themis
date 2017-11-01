package pep

import (
	"fmt"

	"google.golang.org/grpc/naming"
)

type staticResolver struct {
	Name  string
	Addrs []string
}

func newStaticResolver(name string, addrs ...string) naming.Resolver {
	return &staticResolver{Name: name, Addrs: addrs}
}

func (r *staticResolver) Resolve(target string) (naming.Watcher, error) {
	if target != r.Name {
		return nil, fmt.Errorf("%q is an invalid target for resolver %q", target, r.Name)
	}

	return &staticWatcher{Addrs: r.Addrs}, nil
}
