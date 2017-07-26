package yast

import (
	"fmt"
	"strings"

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
		return nil, bindError(err, fmt.Sprintf("%d", i))
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
		return nil, newUnknownEffectError(s, src)
	}

	obls, err := ctx.unmarshalObligations(m)
	if err != nil {
		return nil, bindError(err, src)
	}

	return pdp.NewRule(ID, !ok, target, cond, effect, obls), nil
}

func (ctx context) unmarshalRulesItem(v interface{}, i int) (*pdp.Rule, boundError) {
	m, err := ctx.validateMap(v, "rule")
	if err != nil {
		return nil, bindError(err, fmt.Sprintf("%d", i))
	}

	return ctx.unmarshalRule(m, i)
}

func (ctx context) unmarshalRules(v interface{}) ([]*pdp.Rule, boundError) {
	rules := []*pdp.Rule{}

	items, err := ctx.validateList(v, "policy rules")
	if err != nil {
		return nil, err
	}

	for i, item := range items {
		r, err := ctx.unmarshalRulesItem(item, i+1)
		if err != nil {
			return nil, err
		}

		rules = append(rules, r)
	}

	return rules, nil
}
