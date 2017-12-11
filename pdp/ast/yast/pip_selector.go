package yast

import (
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

func makePIPSelector(ctx context, loc []string, m map[interface{}]interface{}) (pdp.PIPSelector, boundError) {
	items, err := ctx.extractList(m, yastTagPath, "path")
	if err != nil {
		return pdp.PIPSelector{}, bindErrorf(err, "selector(%s.%s)", loc[0], loc[1])
	}

	path := make([]pdp.Expression, len(items))
	for i, item := range items {
		e, err := ctx.unmarshalExpression(item)
		if err != nil {
			return pdp.PIPSelector{}, bindErrorf(bindErrorf(err, "%d", i), "selector(%s.%s)", loc[0], loc[1])
		}

		path[i] = e
	}

	s, err := ctx.extractString(m, yastTagType, "type")
	if err != nil {
		return pdp.PIPSelector{}, bindErrorf(err, "selector(%s.%s)", loc[0], loc[1])
	}

	t, ok := pdp.TypeIDs[strings.ToLower(s)]
	if !ok {
		return pdp.PIPSelector{}, bindErrorf(newUnknownTypeError(s), "selector(%s.%s)", loc[0], loc[1])
	}

	if t == pdp.TypeUndefined {
		return pdp.PIPSelector{}, bindErrorf(newInvalidTypeError(t), "selector(%s.%s)", loc[0], loc[1])
	}

	return pdp.MakePIPSelector(loc[0], loc[1], path, t), nil
}
