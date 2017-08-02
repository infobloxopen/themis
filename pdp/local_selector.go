package pdp

import (
	"fmt"

	"github.com/infobloxopen/go-trees/strtree"
)

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
	v, ok := ctx.c.Get(s.content)
	if !ok {
		return undefinedValue, bindError(newMissingContentError(), s.describe())
	}

	items, ok := v.(*strtree.Tree)
	if !ok {
		panic(fmt.Errorf("Local selector: Invalid content %s (expected *strtree.Tree but got %T)", s.content, v))
	}

	v, ok = items.Get(s.item)
	if !ok {
		return undefinedValue, bindError(newMissingContentItemError(), s.describe())
	}

	item, ok := v.(ContentItem)
	if !ok {
		panic(fmt.Errorf("Local selector: Invalid content item %s.%s (expected ContentItem but got %T)",
			s.content, s.item, v))
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
