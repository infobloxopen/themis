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

func (s *selector) SelectorFunc(uri *url.URL, path []pdp.Expression, t pdp.Type, opts ...pdp.SelectorOption) (pdp.Expression, error) {
	return MakePipSelector(s.pool, uri, path, t, opts...)
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

func (s *selectorUnix) SelectorFunc(uri *url.URL, path []pdp.Expression, t pdp.Type, opts ...pdp.SelectorOption) (pdp.Expression, error) {
	return MakePipSelector(s.pool, uri, path, t, opts...)
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

func (s *selectorK8s) SelectorFunc(uri *url.URL, path []pdp.Expression, t pdp.Type, opts ...pdp.SelectorOption) (pdp.Expression, error) {
	return MakePipSelector(s.pool, uri, path, t, opts...)
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

	def pdp.Expression
	err pdp.Expression
}

// MakePipSelector creates an expression base on PIP selector. Client pool must
// correspond to URI schema.
func MakePipSelector(clients *clientsPool, uri *url.URL, path []pdp.Expression, t pdp.Type, opts ...pdp.SelectorOption) (pdp.Expression, error) {
	ps := PipSelector{
		clients: clients,
		net:     "tcp",
		addr:    uri.Host,
		id:      uri.Path,
		path:    path,
		t:       t,
	}

	for _, opt := range opts {
		switch opt.Name {
		case pdp.SelectorOptionDefault:
			if exp, ok := opt.Data.(pdp.Expression); ok {
				ps.def = exp
			} else {
				panic("bad data provided as pip selector option " + pdp.SelectorOptionDefault)
			}
		case pdp.SelectorOptionError:
			if exp, ok := opt.Data.(pdp.Expression); ok {
				ps.err = exp
			} else {
				panic("bad data provided as pip selector option " + pdp.SelectorOptionError)
			}
		}
	}

	switch strings.ToLower(uri.Scheme) {
	case pipSelectorScheme:

	case pipUnixSelectorScheme:
		ps.net = "unix"
		ps.addr = uri.Path
		ps.id = uri.Fragment

	case pipK8sSelectorScheme:
		ps.k8s = true

	default:
		return PipSelector{}, fmt.Errorf("Unknown pip selector scheme %q", uri.Scheme)
	}

	return ps, nil
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
			return s.handleError(ctx, fmt.Errorf("Failed to calculate argument %d: %s", i+1, err))
		}

		vals = append(vals, v)
	}

	c, err := s.clients.Get(s.addr)
	if err != nil {
		return s.handleError(ctx, fmt.Errorf("Failed to get PIP client for %s: %s", s.addr, err))
	}
	defer s.clients.Free(s.addr)

	r, err := c.Get(s.id, vals)
	if err != nil {
		return s.handleError(ctx, fmt.Errorf("Failed to get information from PIP: %s", err))
	}

	r, err = r.Rebind(s.t)
	if err != nil {
		return s.handleError(ctx, fmt.Errorf("Expected content with value type %q but got %q", s.t, r.GetResultType()))
	}

	return r, nil
}

func (s PipSelector) handleError(ctx *pdp.Context, err error) (pdp.AttributeValue, error) {
	if _, ok := err.(*pdp.MissingValueError); ok && s.def != nil {
		return s.def.Calculate(ctx)
	}

	if s.err != nil {
		return s.err.Calculate(ctx)
	}

	return pdp.UndefinedValue, err
}
