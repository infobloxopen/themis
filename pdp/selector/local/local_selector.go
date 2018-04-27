package local

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

type selector struct{}

func (s *selector) Enabled() bool {
	return true
}

func (s *selector) SelectorFunc(uri string, path []pdp.Expression, t pdp.Type) (pdp.Expression, error) {
	return MakeLocalSelector(uri, path, t)
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
}

// MakeLocalSelector creates instance of local selector. Arguments content and
// item are id of content in storage and id of content item within content.
// Argument path defines set of expressions to get a value of type t. Local
// selector implements late binding and checks path and type on any evaluation.
func MakeLocalSelector(uri string, path []pdp.Expression, t pdp.Type) (pdp.Expression, error) {
	loc := strings.Split(uri, "/")
	if len(loc) != 2 {
		err := fmt.Errorf("Expected selector location in form of <Content-ID>/<Item-ID> got %s", uri)
		return LocalSelector{}, err
	}
	return LocalSelector{
		content: loc[0],
		item:    loc[1],
		path:    path,
		t:       t}, nil
}

// GetResultType implements Expression interface and returns type of final value
// expected by the selector from corresponding content.
func (s LocalSelector) GetResultType() pdp.Type {
	return s.t
}

func typeMismatchError(expected, actual pdp.Type) error {
	return fmt.Errorf("Invalid conent item type. Expected %q but got %q", expected, actual)
}

// Calculate implements Expression interface and returns calculated value
func (s LocalSelector) Calculate(ctx *pdp.Context) (pdp.AttributeValue, error) {
	var r pdp.AttributeValue

	item, err := ctx.GetContentItem(s.content, s.item)
	if err != nil {
		return r, err
	}

	t := item.GetType()
	if s.t != t {
		return r, typeMismatchError(s.t, t)
	}

	r, err = item.Get(s.path, ctx)
	if err != nil {
		return r, err
	}

	t = r.GetResultType()
	if s.t != t {
		return r, typeMismatchError(s.t, t)
	}

	return r, nil
}
