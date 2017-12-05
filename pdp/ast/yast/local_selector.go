package yast

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

func makeLocalSelector(ctx context, loc []string, m map[interface{}]interface{}) (pdp.LocalSelector, boundError) {
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

	res := pdp.MakeLocalSelector(loc[0], loc[1], path, t)

	fmt.Printf("Res is %v\n", res)
	return res, nil
}
