package local

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

const localSelectorScheme = "local"

type selector struct{}

func (s *selector) Scheme() string {
	return localSelectorScheme
}

func (s *selector) Enabled() bool {
	return true
}

func (s *selector) SelectorFunc(uri *url.URL, path []pdp.Expression, t pdp.Type, def, err pdp.Expression) (pdp.Expression, error) {
	return MakeLocalSelector(uri, path, t, def, err)
}

func (s *selector) Initialize() {}

// LocalSelector represent local selector expression. The expression extracts
// value from local content storage by given path and validates that result
// has desired type.
type LocalSelector struct {
	content string
	item    string
	path    []pdp.Expression
	t       pdp.Type
	def     pdp.Expression
	err     pdp.Expression
}

// MakeLocalSelector creates instance of local selector. Arguments content and
// item are id of content in storage and id of content item within content.
// Argument path defines set of expressions to get a value of type t. Local
// selector implements late binding and checks path and type on any evaluation.
// If content storage doesn't have a value for given path the value of
// def expression is returned if it was provided. In case if other error occurs
// the value of err expression is returned if it was provided.
func MakeLocalSelector(uri *url.URL, path []pdp.Expression, t pdp.Type, def, err pdp.Expression) (pdp.Expression, error) {
	loc := strings.Split(uri.Opaque, "/")
	if len(loc) != 2 {
		return nil, fmt.Errorf("Expected selector location in form of <Content-ID>/<Item-ID> got %s", uri)
	}

	return LocalSelector{
		content: loc[0],
		item:    loc[1],
		path:    path,
		t:       t,
		def:     def,
		err:     err,
	}, nil
}

// GetResultType implements Expression interface and returns type of final value
// expected by the selector from corresponding content.
func (s LocalSelector) GetResultType() pdp.Type {
	return s.t
}

// Calculate implements Expression interface and returns calculated value
func (s LocalSelector) Calculate(ctx *pdp.Context) (pdp.AttributeValue, error) {
	item, err := ctx.GetContentItem(s.content, s.item)
	if err != nil {
		return s.handleError(ctx, err)
	}

	r, err := item.Get(s.path, ctx)
	if err != nil {
		return s.handleError(ctx, err)
	}

	r, err = r.Rebind(s.t)
	if err != nil {
		return s.handleError(ctx, fmt.Errorf(
			"Expected content with value type %q but got %q",
			s.t,
			r.GetResultType(),
		))
	}

	return r, nil
}

func (s LocalSelector) handleError(ctx *pdp.Context, err error) (pdp.AttributeValue, error) {
	if _, ok := err.(*pdp.MissingValueError); ok && s.def != nil {
		return s.def.Calculate(ctx)
	}

	if s.err != nil {
		return s.err.Calculate(ctx)
	}

	return pdp.UndefinedValue, err
}
