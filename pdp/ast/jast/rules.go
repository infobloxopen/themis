package jast

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

func (ctx context) unmarshalCondition(m map[interface{}]interface{}) (pdp.Expression, boundError) {
	v, ok := m[yastTagCondition]
	if !ok {
		return nil, nil
	}

	e, err := ctx.unmarshalExpression(v)
	if err != nil {
		return nil, err
	}

	t := e.GetResultType()
	if t != pdp.TypeBoolean {
		return nil, newConditionTypeError(t)
	}

	return e, nil
}

func (ctx context) unmarshalRule(m map[interface{}]interface{}, i int) (*pdp.Rule, boundError) {
	ID, ok, err := ctx.extractStringOpt(m, yastTagID, "id")
	if err != nil {
		return nil, bindErrorf(err, "%d", i)
	}

	src := makeSource("rule", ID, !ok, i)

	target, err := ctx.unmarshalTarget(m)
	if err != nil {
		return nil, bindError(err, src)
	}

	cond, err := ctx.unmarshalCondition(m)
	if err != nil {
		return nil, bindError(err, src)
	}

	s, err := ctx.extractString(m, yastTagEffect, "effect")
	if err != nil {
		return nil, bindError(err, src)
	}

	effect, ok := pdp.EffectIDs[strings.ToLower(s)]
	if !ok {
		return nil, bindError(newUnknownEffectError(s), src)
	}

	obls, err := ctx.unmarshalObligations(m)
	if err != nil {
		return nil, bindError(err, src)
	}

	return pdp.NewRule(ID, !ok, target, cond, effect, obls), nil
}

func (ctx context) unmarshalRuleEntity(m map[interface{}]interface{}, ID string, hidden bool, effect interface{}) (*pdp.Rule, boundError) {
	src := makeSource("rule", ID, hidden, 0)

	target, err := ctx.unmarshalTarget(m)
	if err != nil {
		return nil, bindError(err, src)
	}

	cond, err := ctx.unmarshalCondition(m)
	if err != nil {
		return nil, bindError(err, src)
	}

	s, err := ctx.validateString(effect, "effect")
	if err != nil {
		return nil, bindError(err, src)
	}

	eff, ok := pdp.EffectIDs[strings.ToLower(s)]
	if !ok {
		return nil, bindError(newUnknownEffectError(s), src)
	}

	obls, err := ctx.unmarshalObligations(m)
	if err != nil {
		return nil, bindError(err, src)
	}

	return pdp.NewRule(ID, !ok, target, cond, eff, obls), nil
}

func (ctx context) decodeRuleItem(d *json.Decoder, i int) (*pdp.Rule, error) {
	m, err := ctx.decodeObject(d, "rule")
	if err != nil {
		return nil, err
	}

	return ctx.unmarshalRule(m, i)
}

func (ctx *context) decodeRules(d *json.Decoder) ([]*pdp.Rule, error) {
	err := jparser.CheckArrayStart(d, "rules")
	if err != nil {
		return nil, err
	}

	rules := []*pdp.Rule{}

	err = jparser.UnmarshalObjectArray(d, func(idx int, d *json.Decoder) error {
		e, err := ctx.decodeRuleItem(d, idx+1)
		if err != nil {
			return bindErrorf(err, "%d", idx)
		}

		rules = append(rules, e)

		return nil
	}, "rules")
	if err != nil {
		return nil, err
	}

	return rules, nil
}
