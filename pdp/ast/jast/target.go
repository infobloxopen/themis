package jast

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

func (ctx context) unmarshalAdjustedArguments(val pdp.Expression, attr pdp.Expression, d *json.Decoder) (pdp.Expression, pdp.Expression, error) {
	e, err := ctx.unmarshalExpression(d)
	if err != nil {
		return nil, nil, err
	}

	switch a := e.(type) {
	case pdp.AttributeValue:
		if val != nil {
			return nil, nil, newMatchFunctionBothValuesError()
		}

		return a, attr, nil

	case pdp.LocalSelector:
		if val != nil {
			return nil, nil, newMatchFunctionBothValuesError()
		}

		return a, attr, nil

	case pdp.AttributeDesignator:
		if attr != nil {
			return nil, nil, newMatchFunctionBothAttrsError()
		}

		return val, a, nil
	}

	return nil, nil, newInvalidMatchFunctionArgError(e)
}

func (ctx context) unmarshalAdjustedArgumentPair(d *json.Decoder) (pdp.Expression, pdp.Expression, error) {
	if err := jparser.CheckArrayStart(d, "function arguments"); err != nil {
		return nil, nil, err
	}

	var (
		first, second pdp.Expression
		numArgs       int
	)

	if err := jparser.UnmarshalObjectArray(d, func(idx int, d *json.Decoder) error {
		var err error

		switch idx {
		case 1:
			numArgs++
			first, second, err = ctx.unmarshalAdjustedArguments(nil, nil, d)
			return err
		case 2:
			numArgs++
			first, second, err = ctx.unmarshalAdjustedArguments(first, second, d)
			return err
		}

		return nil
	}, "function arguments"); err != nil {
		return nil, nil, err
	}

	if numArgs != 2 {
		return nil, nil, newMatchFunctionArgsNumberError(numArgs)
	}

	return first, second, nil
}

func (ctx context) unmarshalTargetMatchExpression(id string, d *json.Decoder) (pdp.Expression, error) {
	typeFunctionMap, ok := pdp.TargetCompatibleExpressions[strings.ToLower(id)]
	if !ok {
		return nil, newUnknownMatchFunctionError(id)
	}

	first, second, err := ctx.unmarshalAdjustedArgumentPair(d)
	if err != nil {
		return nil, bindError(err, id)
	}

	firstType := first.GetResultType()
	secondType := second.GetResultType()

	subTypeFunctionMap, ok := typeFunctionMap[firstType]
	if !ok {
		return nil, newMatchFunctionCastError(id, firstType, secondType)
	}

	maker, ok := subTypeFunctionMap[secondType]
	if !ok {
		return nil, newMatchFunctionCastError(id, firstType, secondType)
	}

	return maker(first, second), nil
}

func (ctx context) unmarshalTargetAllOfItem(d *json.Decoder) (pdp.Match, error) {
	m := pdp.Match{}
	var exp pdp.Expression

	if err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		var err error
		exp, err = ctx.unmarshalTargetMatchExpression(k, d)
		if err != nil {
			return bindError(err, k)
		}

		return nil
	}, "function identifier"); err != nil {
		return m, err
	}

	return pdp.MakeMatch(exp), nil
}

func (ctx context) unmarshalTargetAnyOfItem(d *json.Decoder) (pdp.AllOf, error) {
	all := pdp.MakeAllOf()

	if err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		if strings.ToLower(k) == yastTagAll {
			if err := jparser.CheckArrayStart(d, "list of all expressions"); err != nil {
				return err
			}

			if err := jparser.UnmarshalObjectArray(d, func(idx int, d *json.Decoder) error {
				m, err := ctx.unmarshalTargetAllOfItem(d)
				if err != nil {
					return bindError(bindErrorf(err, "%d", idx), k)
				}

				all.Append(m)

				return nil
			}, "list of all expressions"); err != nil {
				return err
			}
		} else {
			e, err := ctx.unmarshalTargetMatchExpression(k, d)
			if err != nil {
				return err
			}

			m := pdp.MakeMatch(e)
			all.Append(m)
		}

		return nil
	}, "function identifier"); err != nil {
		return all, err
	}

	return all, nil
}

func (ctx context) unmarshalTargetItem(d *json.Decoder) (pdp.AnyOf, error) {
	any := pdp.MakeAnyOf()

	if err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		if strings.ToLower(k) == yastTagAny {
			if err := jparser.CheckArrayStart(d, "list of any expressions"); err != nil {
				return err
			}

			if err := jparser.UnmarshalObjectArray(d, func(idx int, d *json.Decoder) error {
				all, err := ctx.unmarshalTargetAnyOfItem(d)
				if err != nil {
					return bindError(bindErrorf(err, "%d", idx), k)
				}

				any.Append(all)

				return nil
			}, "list of any expressions"); err != nil {
				return err
			}
		} else {
			e, err := ctx.unmarshalTargetMatchExpression(k, d)
			if err != nil {
				return err
			}

			all := pdp.MakeAllOf()
			all.Append(pdp.MakeMatch(e))
			any.Append(all)
		}

		return nil
	}, "function identifier"); err != nil {
		return any, err
	}

	return any, nil
}

func (ctx *context) unmarshalTarget(d *json.Decoder) (pdp.Target, error) {
	t := pdp.MakeTarget()
	if err := jparser.CheckArrayStart(d, "target"); err != nil {
		return t, err
	}

	if err := jparser.UnmarshalObjectArray(d, func(idx int, d *json.Decoder) error {
		item, err := ctx.unmarshalTargetItem(d)
		if err != nil {
			return bindErrorf(bindErrorf(err, "%d", idx), "target")
		}

		t.Append(item)

		return nil
	}, "target"); err != nil {
		return t, err
	}

	return t, nil
}
