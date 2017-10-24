package jast

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

func (ctx *context) decodeAttributeDeclarations(d *json.Decoder) boundError {
	err := jparser.CheckObjectStart(d, "attribute declarations")
	if err != nil {
		return bindError(err, yastTagAttributes)
	}

	if err = jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		tstr, err := jparser.GetString(d, "attribute data type")
		if err != nil {
			return err
		}

		t, ok := pdp.TypeIDs[strings.ToLower(tstr)]
		if !ok {
			return bindError(newAttributeTypeError(tstr), k)
		}

		ctx.attrs[k] = pdp.MakeAttribute(k, t)

		return nil
	}, "attribute declarations"); err != nil {
		return bindError(err, yastTagAttributes)
	}

	return nil
}
