package yast

import (
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

func (ctx context) unmarshalSelector(v interface{}) (pdp.Expression, boundError) {
	m, err := ctx.validateMap(v, "selector attributes")
	if err != nil {
		return nil, err
	}

	uri, err := ctx.extractString(m, yastTagURI, "selector URI")
	if err != nil {
		return nil, err
	}

	id, ierr := url.Parse(uri)
	if ierr != nil {
		return nil, newSelectorURIError(uri, ierr)
	}

	items, err := ctx.extractList(m, yastTagPath, "path")
	if err != nil {
		return nil, bindErrorf(err, "selector(%s)", uri)
	}

	path := make([]pdp.Expression, len(items))
	for i, item := range items {
		e, err := ctx.unmarshalExpression(item)
		if err != nil {
			return nil, bindErrorf(bindErrorf(err, "%d", i), "selector(%s)", uri)
		}

		path[i] = e
	}

	st, err := ctx.extractString(m, yastTagType, "type")
	if err != nil {
		return nil, bindErrorf(err, "selector(%s)", uri)
	}

	t := ctx.symbols.GetType(st)
	if t == nil {
		return nil, bindErrorf(newUnknownTypeError(st), "selector(%s)", uri)
	}

	if t == pdp.TypeUndefined {
		return nil, bindErrorf(newInvalidTypeError(t), "selector(%s)", uri)
	}

	var opts []pdp.SelectorOption

	defMap, ok, err := ctx.extractMapOpt(m, yastTagDefault, "default")
	if err != nil {
		return nil, bindErrorf(err, "selector(%s).default", uri)
	}
	if ok {
		defExp, err := ctx.unmarshalExpression(defMap)
		if err != nil {
			return nil, bindErrorf(err, "selector(%s).default", uri)
		}
		if defExp != nil {
			if defExp.GetResultType() != t {
				return nil, bindErrorf(newInvalidTypeError(t), "selector(%s).default", uri)
			}
			opts = append(opts, pdp.SelectorOption{Name: pdp.SelectorOptionDefault, Data: defExp})
		}
	}

	errMap, ok, err := ctx.extractMapOpt(m, yastTagError, "error")
	if err != nil {
		return nil, bindErrorf(err, "selector(%s).error", uri)
	}
	if ok {
		errExp, err := ctx.unmarshalExpression(errMap)
		if err != nil {
			return nil, bindErrorf(err, "selector(%s).error", uri)
		}
		if errExp != nil {
			if errExp.GetResultType() != t {
				return nil, bindErrorf(newInvalidTypeError(t), "selector(%s).error", uri)
			}
			opts = append(opts, pdp.SelectorOption{Name: pdp.SelectorOptionError, Data: errExp})
		}
	}

	aggStr, ok, err := ctx.extractStringOpt(m, yastTagAggregation, "aggregation")
	if err != nil {
		return nil, bindErrorf(err, "selector(%s)", uri)
	}

	if ok && aggStr != "" {
		a, ok := pdp.AggTypeIDs[strings.ToLower(aggStr)]
		if !ok {
			return nil, bindErrorf(newUnknownAggregationTypeError(aggStr), "selector(%s).aggregation", uri)
		}
		if (a == pdp.AggTypeAppend || a == pdp.AggTypeAppendUnique) && t != pdp.TypeListOfStrings {
			return nil, bindErrorf(newInvalidAggregationTypeError(aggStr, t), "selector(%s).aggregation", uri)
		}
		opts = append(opts, pdp.SelectorOption{Name: pdp.SelectorOptionAggregation, Data: a})
	}

	e, eErr := pdp.MakeSelector(id, path, t, opts...)
	if eErr != nil {
		return nil, bindErrorf(eErr, "selector(%s)", uri)
	}
	return e, nil
}
