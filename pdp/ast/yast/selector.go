package yast

import (
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pdp/selector"
)

func (ctx context) unmarshalSelector(v interface{}) (pdp.Expression, boundError) {
	var ret pdp.Expression

	m, err := ctx.validateMap(v, "selector attributes")
	if err != nil {
		return ret, err
	}

	uri, err := ctx.extractString(m, yastTagURI, "selector URI")
	if err != nil {
		return ret, err
	}

	id, ierr := url.Parse(uri)
	if ierr != nil {
		return ret, newSelectorURIError(uri, ierr)
	}

	items, err := ctx.extractList(m, yastTagPath, "path")
	if err != nil {
		return ret, bindErrorf(err, "selector(%s)", uri)
	}

	path := make([]pdp.Expression, len(items))
	for i, item := range items {
		e, err := ctx.unmarshalExpression(item)
		if err != nil {
			return ret, bindErrorf(bindErrorf(err, "%d", i), "selector(%s)", uri)
		}

		path[i] = e
	}

	st, err := ctx.extractString(m, yastTagType, "type")
	if err != nil {
		return ret, bindErrorf(err, "selector(%s)", uri)
	}

	t, ok := pdp.BuiltinTypeIDs[strings.ToLower(st)]
	if !ok {
		return ret, bindErrorf(newUnknownTypeError(uri), "selector(%s)", uri)
	}

	if t == pdp.TypeUndefined {
		return ret, bindErrorf(newInvalidTypeError(t), "selector(%s)", uri)
	}

	var e error
	ret, e = selector.MakeSelector(strings.ToLower(id.Scheme), id.Opaque, path, t)
	if e != nil {
		return ret, bindErrorf(e, "selector(%s)", uri)
	}
	return ret, nil
}
