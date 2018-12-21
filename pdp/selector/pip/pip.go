package pip

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

const (
	pipSelectorScheme     = "pip"
	pipUnixSelectorScheme = "pip+unix"
	pipK8sSelectorScheme  = "pip+k8s"
)

type selector struct {
	pool *clientsPool
}

func (s *selector) Scheme() string {
	return pipSelectorScheme
}

func (s *selector) Enabled() bool {
	return true
}

func (s *selector) SelectorFunc(uri *url.URL, path []pdp.Expression, t pdp.Type) (pdp.Expression, error) {
	return MakePipSelector(s.pool, uri, path, t)
}

func (s *selector) Initialize() {
	s.pool = NewTCPClientsPool()
	go s.pool.cleaner(nil, make(chan struct{}))
}

type selectorUnix struct {
	pool *clientsPool
}

func (s *selectorUnix) Scheme() string {
	return pipUnixSelectorScheme
}

func (s *selectorUnix) Enabled() bool {
	return true
}

func (s *selectorUnix) SelectorFunc(uri *url.URL, path []pdp.Expression, t pdp.Type) (pdp.Expression, error) {
	return MakePipSelector(s.pool, uri, path, t)
}

func (s *selectorUnix) Initialize() {
	s.pool = NewUnixClientsPool()
	go s.pool.cleaner(nil, make(chan struct{}))
}

type selectorK8s struct {
	pool *clientsPool
}

func (s *selectorK8s) Scheme() string {
	return pipK8sSelectorScheme
}

func (s *selectorK8s) Enabled() bool {
	return true
}

func (s *selectorK8s) SelectorFunc(uri *url.URL, path []pdp.Expression, t pdp.Type) (pdp.Expression, error) {
	return MakePipSelector(s.pool, uri, path, t)
}

func (s *selectorK8s) Initialize() {
	s.pool = NewK8sClientsPool()
	go s.pool.cleaner(nil, make(chan struct{}))
}

// PipSelector represents selector for Unified PIP.
type PipSelector struct {
	clients *clientsPool

	net  string
	k8s  bool
	addr string
	id   string

	path []pdp.Expression
	t    pdp.Type
}

// MakePipSelector creates an expression base on PIP selector. Client pool must
// correspond to URI schema.
func MakePipSelector(clients *clientsPool, uri *url.URL, path []pdp.Expression, t pdp.Type) (pdp.Expression, error) {
	switch strings.ToLower(uri.Scheme) {
	case pipSelectorScheme:
		return PipSelector{
			clients: clients,
			net:     "tcp",
			addr:    uri.Host,
			id:      uri.Path,
			path:    path,
			t:       t,
		}, nil

	case pipUnixSelectorScheme:
		return PipSelector{
			clients: clients,
			net:     "unix",
			addr:    uri.Path,
			id:      uri.Fragment,
			path:    path,
			t:       t,
		}, nil

	case pipK8sSelectorScheme:
		return PipSelector{
			clients: clients,
			net:     "tcp",
			k8s:     true,
			addr:    uri.Host,
			id:      uri.Path,
			path:    path,
			t:       t,
		}, nil
	}

	return PipSelector{}, fmt.Errorf("Unknown pip selector scheme %q", uri.Scheme)
}

// GetResultType implements pdp.Expression interface and returns type of
// selector's result.
func (s PipSelector) GetResultType() pdp.Type {
	return s.t
}

// Calculate implements pdp.Expression interface and obtains result from
// unified PIP for given context.
func (s PipSelector) Calculate(ctx *pdp.Context) (pdp.AttributeValue, error) {
	vals := make([]pdp.AttributeValue, 0, len(s.path))
	for i, item := range s.path {
		v, err := item.Calculate(ctx)
		if err != nil {
			return pdp.UndefinedValue, fmt.Errorf("Failed to calculate argument %d: %s", i+1, err)
		}

		vals = append(vals, v)
	}

	c, err := s.clients.Get(s.addr)
	if err != nil {
		return pdp.UndefinedValue, fmt.Errorf("Failed to get PIP client for %s: %s", s.addr, err)
	}
	defer s.clients.Free(s.addr)

	r, err := c.Get(s.id, vals)
	if err != nil {
		return pdp.UndefinedValue, fmt.Errorf("Failed to get information from PIP: %s", err)
	}

	r, err = r.Rebind(s.t)
	if err != nil {
		return pdp.UndefinedValue, fmt.Errorf("Expected content with value type %q but got %q", s.t, r.GetResultType())

	}

	return r, nil
}
