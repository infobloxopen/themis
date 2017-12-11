package jast

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

func (ctx context) unmarshalSelector(d *json.Decoder) (pdp.Expression, error) {
	if err := jparser.CheckObjectStart(d, "selector"); err != nil {
		return pdp.LocalSelector{}, err
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
		return pdp.LocalSelector{}, err
	}

	id, err := url.Parse(uri)
	if err != nil {
		return pdp.LocalSelector{}, newSelectorURIError(uri, err)
	}

	scheme := strings.ToLower(id.Scheme)
	if scheme == "local" {
		loc := strings.Split(id.Opaque, "/")
		if len(loc) != 2 {
			return pdp.LocalSelector{}, newSelectorLocationError(id.Opaque, uri)
		}

		t, ok := pdp.TypeIDs[strings.ToLower(st)]
		if !ok {
			return pdp.LocalSelector{}, bindErrorf(newUnknownTypeError(uri), "selector(%s.%s)", loc[0], loc[1])
		}

		if t == pdp.TypeUndefined {
			return pdp.LocalSelector{}, bindErrorf(newInvalidTypeError(t), "selector(%s.%s)", loc[0], loc[1])
		}

		return pdp.MakeLocalSelector(loc[0], loc[1], path, t), nil
	} else if strings.ToLower(id.Scheme) == "pip" {
		loc := strings.Split(id.Opaque, "/")
		if len(loc) != 2 {
			return pdp.PIPSelector{}, newSelectorLocationError(id.Opaque, uri)
		}

		t, ok := pdp.TypeIDs[strings.ToLower(st)]
		if !ok {
			return pdp.PIPSelector{}, bindErrorf(newUnknownTypeError(uri), "selector(%s.%s)", loc[0], loc[1])
		}

		if t == pdp.TypeUndefined {
			return pdp.PIPSelector{}, bindErrorf(newInvalidTypeError(t), "selector(%s.%s)", loc[0], loc[1])
		}

		return pdp.MakePIPSelector(loc[0], loc[1], path, t), nil
	}

	return pdp.LocalSelector{}, newUnsupportedSelectorSchemeError(id.Scheme, uri)
}
