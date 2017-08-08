package pdp

import "fmt"

type LocalSelector struct {
	content string
	item    string
	path    []Expression
	t       int
}

func MakeLocalSelector(content, item string, path []Expression, t int) LocalSelector {
	return LocalSelector{
		content: content,
		item:    item,
		path:    path,
		t:       t}
}

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

	r, err := item.get(s.path, ctx)
	if err != nil {
		return undefinedValue, bindError(err, s.describe())
	}

	t := r.GetResultType()
	if t != s.t {
		return undefinedValue, bindError(newInvalidContentItemTypeError(s.t, t), s.describe())
	}

	return r, nil
}
