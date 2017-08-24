package pdp

import "fmt"

type PIPSelector struct {
	content string
	item    string
	path    []Expression
	t       int
}

func MakePIPSelector(content, item string, path []Expression, t int) PIPSelector {
	return PIPSelector{
		content: content,
		item:    item,
		path:    path,
		t:       t}
}

func (s PIPSelector) GetResultType() int {
	return s.t
}

func (s PIPSelector) describe() string {
	return fmt.Sprintf("selector(%s.%s)", s.content, s.item)
}

func (s PIPSelector) calculate(ctx *Context) (AttributeValue, error) {
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
