package yast

import (
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/pdp"
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
		return ret, bindErrorf(err, "selector(%s)", id.Opaque)
	}

	path := make([]pdp.Expression, len(items))
	for i, item := range items {
		e, err := ctx.unmarshalExpression(item)
		if err != nil {
			return ret, bindErrorf(bindErrorf(err, "%d", i), "selector(%s)", id.Opaque)
		}

		path[i] = e
	}

	st, err := ctx.extractString(m, yastTagType, "type")
	if err != nil {
		return ret, bindErrorf(err, "selector(%s)", id.Opaque)
	}

	t, ok := pdp.TypeIDs[strings.ToLower(st)]
	if !ok {
		return ret, bindErrorf(newUnknownTypeError(uri), "selector(%s)", id.Opaque)
	}

	if t == pdp.TypeUndefined {
		return ret, bindErrorf(newInvalidTypeError(t), "selector(%s)", id.Opaque)
	}

	switch strings.ToLower(id.Scheme) {
	case "local":
		loc := strings.Split(id.Opaque, "/")
		if len(loc) != 2 {
			return ret, newSelectorLocationError(id.Opaque, uri)
		}
		ret = pdp.MakeLocalSelector(loc[0], loc[1], path, t)
		return ret, nil
	case "pip":
		ret = pdp.MakePipSelector(id.Opaque, path, t)
		return ret, nil
	}

	return ret, newUnsupportedSelectorSchemeError(id.Scheme, uri)
}
