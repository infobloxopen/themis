package jast

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pdp/selector"
)

func (ctx context) unmarshalSelector(d *json.Decoder) (pdp.Expression, error) {
	var ret pdp.Expression

	if err := jparser.CheckObjectStart(d, "selector"); err != nil {
		return ret, err
	}

	var (
		uri  string
		path []pdp.Expression
		st   string
	)

	if err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		var err error

		switch strings.ToLower(k) {
		case yastTagURI:
			uri, err = jparser.GetString(d, "selector URI")
			return err

		case yastTagPath:
			if err = jparser.CheckArrayStart(d, "selector path"); err != nil {
				return err
			}

			path = []pdp.Expression{}
			if err = jparser.UnmarshalObjectArray(d, func(idx int, d *json.Decoder) error {
				e, err := ctx.unmarshalExpression(d)
				if err != nil {
					return bindError(bindErrorf(err, "%d", idx), "selector path")
				}

				path = append(path, e)

				return nil
			}, "selector path"); err != nil {
				return err
			}

			return nil

		case yastTagType:
			st, err = jparser.GetString(d, "selector type")
			if err != nil {
				return err
			}

			return nil
		}

		return newUnknownAttributeError(k)
	}, "selector"); err != nil {
		return ret, err
	}

	id, err := url.Parse(uri)
	if err != nil {
		return ret, newSelectorURIError(uri, err)
	}

	t, ok := pdp.BuiltinTypes[strings.ToLower(st)]
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
