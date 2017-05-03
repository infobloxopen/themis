package pdp

import "strings"

func (ctx *yastCtx) unmarshalRuleEffect(m map[interface{}]interface{}) (int, error) {
	s, err := ctx.extractString(m, yastTagEffect, "rule effect")
	if err != nil {
		return EffectIndeterminate, err
	}

	ctx.pushNodeSpec(yastTagEffect)
	defer ctx.popNodeSpec()

	e, ok := EffectIDs[strings.ToLower(s)]
	if !ok {
		return EffectIndeterminate, ctx.errorf("Expected rule effect but got %s", s)
	}

	return e, nil
}

func (ctx *yastCtx) unmarshalCondition(m map[interface{}]interface{}) (ExpressionType, error) {
	v, ok := m[yastTagCondition]
	if !ok {
		return nil, nil
	}

	ctx.pushNodeSpec(yastTagCondition)
	defer ctx.popNodeSpec()

	e, err := ctx.unmarshalExpression(v)
	if err != nil {
		return nil, err
	}

	t := e.getResultType()
	if t != DataTypeBoolean {
		return nil, ctx.errorf("Expected expression with %s result type but got %s",
			DataTypeNames[DataTypeBoolean], DataTypeNames[t])
	}

	return e, nil
}

func (ctx *yastCtx) unmarshalRule(m map[interface{}]interface{}) (RuleType, error) {
	r := RuleType{}

	ID, err := ctx.extractString(m, yastTagID, "rule id")
	if err != nil {
		return r, err
	}

	ctx.pushNodeSpec("%#v", ID)
	defer ctx.popNodeSpec()

	t, err := ctx.unmarshalTarget(m)
	if err != nil {
		return r, err
	}

	c, err := ctx.unmarshalCondition(m)
	if err != nil {
		return r, err
	}

	o, err := ctx.unmarshalObligation(m)
	if err != nil {
		return r, err
	}

	e, err := ctx.unmarshalRuleEffect(m)
	if err != nil {
		return r, err
	}

	r.ID = ID
	r.Target = t
	r.Condition = c
	r.Obligations = o
	r.Effect = e

	return r, nil
}

func (ctx *yastCtx) unmarshalRulesItem(v interface{}, i int, rules []RuleType) ([]RuleType, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	m, err := ctx.validateMap(v, "policy rule")
	if err != nil {
		return nil, err
	}

	r, err := ctx.unmarshalRule(m)
	if err != nil {
		return nil, err
	}

	return append(rules, r), nil
}

func (ctx *yastCtx) unmarshalRules(v interface{}) ([]RuleType, error) {
	ctx.pushNodeSpec(yastTagRules)
	defer ctx.popNodeSpec()

	r := make([]RuleType, 0)

	items, err := ctx.validateList(v, "policy rules")
	if err != nil {
		return nil, err
	}

	for i, item := range items {
		r, err = ctx.unmarshalRulesItem(item, i, r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (ctx *yastCtx) extractRuleCombiningAlg(m map[interface{}]interface{}) (string, RuleCombiningAlgType, map[interface{}]interface{}, error) {
	s, algMap, err := ctx.extractStringOrMapDef(m, yastTagAlg, yastTagDefaultAlg, nil, "rule combining algorithm")
	if err != nil {
		return "", nil, nil, err
	}

	if algMap != nil {
		s, err = ctx.extractString(algMap, yastTagID, "rule combining algorithm id")
		if err != nil {
			return "", nil, nil, err
		}
	}

	ID := strings.ToLower(s)
	a, ok := RuleCombiningAlgMap[ID]
	if !ok {
		return "", nil, nil, ctx.errorf("Excpected policy combining algorithm but got %s", s)
	}

	return ID, a, algMap, nil
}
