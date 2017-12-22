package jast

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
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
