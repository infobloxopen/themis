package pdp

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/go-trees/strtree"
)

type localSelector struct {
	content string
	item    string
	path    []expression
}

func (s localSelector) describe() string {
	return fmt.Sprintf("selector(%s.%s)", s.content, s.item)
}

func (s localSelector) calculate(ctx *Context) (attributeValue, error) {
	v, ok := ctx.c.Get(s.content)
	if !ok {
		return undefinedValue, newMissingContentError(s.describe())
	}

	items, ok := v.(*strtree.Tree)
	if !ok {
		panic(fmt.Errorf("Local selector: Invalid content %s (expected *strtree.Tree but got %T)", s.content, v))
	}

	v, ok = items.Get(s.item)
	if !ok {
		return undefinedValue, newMissingContentItemError(s.describe())
	}

	item, ok := v.(contentItem)
	if !ok {
		panic(fmt.Errorf("Local selector: Invalid content item %s.%s (expected ContentItem but git %T)",
			s.content, s.item, v))
	}

	subItem := item.r

	path := []string{""}
	for _, e := range s.path {
		key, err := e.calculate(ctx)
		if err != nil {
			return undefinedValue, bindError(bindError(err, strings.Join(path, "/")), s.describe())
		}

		path = append(path, key.describe())

		subItem, err = subItem.next(key)
		if err != nil {
			return undefinedValue, bindError(bindError(err, strings.Join(path, "/")), s.describe())
		}
	}

	r, err := subItem.getValue(item.t)
	if err != nil {
		if len(s.path) > 0 {
			err = bindError(err, strings.Join(path, "/"))
		}
		return undefinedValue, bindError(err, s.describe())
	}

	return r, nil
}
