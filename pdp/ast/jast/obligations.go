package jast

import (
	"encoding/json"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pdp/jcon"
)

func (ctx context) unmarshalObligationItem(v interface{}) (pdp.AttributeAssignmentExpression, boundError) {
	m, err := ctx.validateMap(v, "obligation")
	if err != nil {
		return pdp.AttributeAssignmentExpression{}, err
	}

	k, v, err := ctx.getSingleMapPair(m, "obligation")
	if err != nil {
		return pdp.AttributeAssignmentExpression{}, err
	}

	ID, err := ctx.validateString(k, "obligation attribute id")
	if err != nil {
		return pdp.AttributeAssignmentExpression{}, err
	}

	a, ok := ctx.attrs[ID]
	if !ok {
		return pdp.AttributeAssignmentExpression{}, newUnknownAttributeError(ID)
	}

	e, err := ctx.unmarshalValueByType(a.GetType(), v)
	if err != nil {
		return pdp.AttributeAssignmentExpression{}, bindError(err, ID)
	}

	return pdp.MakeAttributeAssignmentExpression(a, e), nil
}

func (ctx context) unmarshalObligations(m map[interface{}]interface{}) ([]pdp.AttributeAssignmentExpression, boundError) {
	items, ok, err := ctx.extractListOpt(m, yastTagObligation, "obligations")
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, nil
	}

	r := []pdp.AttributeAssignmentExpression{}
	for i, item := range items {
		o, err := ctx.unmarshalObligationItem(item)
		if err != nil {
			return nil, bindError(bindErrorf(err, "%d", i), "obligations")
		}

		r = append(r, o)
	}

	return r, nil
}

func (ctx *context) decodeObligations(d *json.Decoder) ([]pdp.AttributeAssignmentExpression, error) {
	if err := jcon.CheckArrayStart(d, "obligations"); err != nil {
		return nil, err
	}

	v, err := ctx.decodeArray(d, "obligations")
	if err != nil {
		return nil, err
	}

	m := map[interface{}]interface{}{yastTagObligation: v}
	return ctx.unmarshalObligations(m)
}
