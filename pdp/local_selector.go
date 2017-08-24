package pdp

import (
	"fmt"
)

// LocalSelector represent local selector expression. The expression extracts
// value from local content storage by given path and validates that result
// has desired type.
type LocalSelector struct {
	content string
	item    string
	path    []Expression
	t       int
}

// MakeLocalSelector creates instance of local selector. Arguments content and
// item are id of content in storage and id of content item within content.
// Argument path defines set of expressions to get a value of type t. Local
// selector implements late binding and checks path and type on any evaluation.
func MakeLocalSelector(content, item string, path []Expression, t int) LocalSelector {
	return LocalSelector{
		content: content,
		item:    item,
		path:    path,
		t:       t}
}

// GetResultType implements Expression interface and returns type of final value
// expected by the selector from corresponding content.
func (s LocalSelector) GetResultType() int {
	return s.t
}

func (s LocalSelector) describe() string {
	return fmt.Sprintf("selector(%s.%s)", s.content, s.item)
}

func (s LocalSelector) calculate(ctx *Context) (AttributeValue, error) {
	item, err := ctx.getContentItem(s.content, s.item)
	if err != nil {
		return undefinedValue, bindError(err, s.describe())
	}

	if s.t != item.t {
		return undefinedValue, bindError(newInvalidContentItemTypeError(s.t, item.t), s.describe())
	}

	r, err := item.Get(s.path, ctx)
	if err != nil {
		return undefinedValue, bindError(err, s.describe())
	}

	t := r.GetResultType()
	if t != s.t {
		return undefinedValue, bindError(newInvalidContentItemTypeError(s.t, t), s.describe())
	}

	return r, nil
}
