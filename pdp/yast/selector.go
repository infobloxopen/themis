package yast

import (
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

func (ctx context) unmarshalSelector(v interface{}) (pdp.LocalSelector, boundError) {
	m, err := ctx.validateMap(v, "selector attributes")
	if err != nil {
		return pdp.LocalSelector{}, err
	}

	s, err := ctx.extractString(m, yastTagURI, "selector URI")
	if err != nil {
		return pdp.LocalSelector{}, err
	}

	ID, ierr := url.Parse(s)
	if ierr != nil {
		return pdp.LocalSelector{}, newSelectorURIError(s, ierr)
	}

	if strings.ToLower(ID.Scheme) == "local" {
		loc := strings.Split(ID.Opaque, "/")
		if len(loc) != 2 {
			return pdp.LocalSelector{}, newSelectorLocationError(ID.Opaque, s)
		}

		items, err := ctx.extractList(m, yastTagPath, "path")
		if err != nil {
			return pdp.LocalSelector{}, bindErrorf(err, "selector(%s.%s)", loc[0], loc[1])
		}

		path := make([]pdp.Expression, len(items))
		for i, item := range items {
			e, err := ctx.unmarshalExpression(item)
			if err != nil {
				return pdp.LocalSelector{}, bindErrorf(bindErrorf(err, "%d", i), "selector(%s.%s)", loc[0], loc[1])
			}

			path[i] = e
		}

		s, err := ctx.extractString(m, yastTagType, "type")
		if err != nil {
			return pdp.LocalSelector{}, bindErrorf(err, "selector(%s.%s)", loc[0], loc[1])
		}

		t, ok := pdp.TypeIDs[strings.ToLower(s)]
		if !ok {
			return pdp.LocalSelector{}, bindErrorf(newUnknownTypeError(s), "selector(%s.%s)", loc[0], loc[1])
		}

		if t == pdp.TypeUndefined {
			return pdp.LocalSelector{}, bindErrorf(newInvalidTypeError(t), "selector(%s.%s)", loc[0], loc[1])
		}

		return pdp.MakeLocalSelector(loc[0], loc[1], path, t), nil
	}

	return pdp.LocalSelector{}, newUnsupportedSelectorSchemeError(ID.Scheme, s)
}
