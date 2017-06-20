package pdp

import "fmt"

type MapperRCAParams struct {
	Argument    ExpressionType
	RulesMap    map[string]*RuleType
	DefaultRule *RuleType
	ErrorRule   *RuleType
	SubAlg      RuleCombiningAlgType
	AlgParams   interface{}
}

func calculateErrorRule(rule *RuleType, ctx *Context, err error) ResponseType {
	if rule != nil {
		return rule.calculate(ctx)
	}

	return ResponseType{EffectIndeterminate, fmt.Sprintf("Mapper Rule Combining Algorithm: %s", err), nil}
}

func getSetOfIDs(v AttributeValueType) ([]string, error) {
	ID, err := ExtractStringValue(v, "argument")
	if err == nil {
		return []string{ID}, nil
	}

	setIDs, err := ExtractSetOfStringsValue(v, "argument")
	if err == nil {
		return sortSetOfStrings(setIDs), nil
	}

	listIDs, err := ExtractListOfStringsValue(v, "argument")
	if err == nil {
		return listIDs, nil
	}

	return nil, fmt.Errorf("Expected %s, %s or %s as argument but got %s",
		DataTypeNames[DataTypeString], DataTypeNames[DataTypeSetOfStrings], DataTypeNames[DataTypeListOfStrings], DataTypeNames[v.DataType])
}

func getRulesMap(rules []*RuleType, params *MapperRCAParams) map[string]*RuleType {
	if params.RulesMap != nil {
		return params.RulesMap
	}

	m := make(map[string]*RuleType)
	for _, rule := range rules {
		m[rule.ID] = rule
	}

	return m
}

func collectSubRules(IDs []string, m map[string]*RuleType) []*RuleType {
	rules := []*RuleType{}
	for _, ID := range IDs {
		rule, ok := m[ID]
		if ok {
			rules = append(rules, rule)
		}
	}

	return rules
}

func MapperRCA(rules []*RuleType, params interface{}, ctx *Context) ResponseType {
	mapperParams := params.(*MapperRCAParams)

	v, err := mapperParams.Argument.calculate(ctx)
	if err != nil {
		switch err.(type) {
		case MissingValueError, *MissingValueError:
			if mapperParams.DefaultRule != nil {
				return mapperParams.DefaultRule.calculate(ctx)
			}
		}

		return calculateErrorRule(mapperParams.ErrorRule, ctx, err)
	}

	if mapperParams.SubAlg != nil {
		IDs, err := getSetOfIDs(v)
		if err != nil {
			return calculateErrorRule(mapperParams.ErrorRule, ctx, err)
		}

		r := mapperParams.SubAlg(collectSubRules(IDs, getRulesMap(rules, mapperParams)), mapperParams.AlgParams, ctx)
		if r.Effect == EffectNotApplicable && mapperParams.DefaultRule != nil {
			return mapperParams.DefaultRule.calculate(ctx)
		}

		return r
	}

	ID, err := ExtractStringValue(v, "argument")
	if err != nil {
		return calculateErrorRule(mapperParams.ErrorRule, ctx, err)
	}

	if mapperParams.RulesMap != nil {
		rule, ok := mapperParams.RulesMap[ID]
		if ok {
			return rule.calculate(ctx)
		}
	} else {
		for _, rule := range rules {
			if rule.ID == ID {
				return rule.calculate(ctx)
			}
		}
	}

	if mapperParams.DefaultRule != nil {
		return mapperParams.DefaultRule.calculate(ctx)
	}

	return ResponseType{EffectNotApplicable, "Ok", nil}
}
