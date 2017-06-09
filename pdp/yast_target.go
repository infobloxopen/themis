package pdp

import (
	"fmt"
	"strings"
)

type twoArgumentsFunctionType func(first ExpressionType, second ExpressionType) ExpressionType

var targetCompatibleExpressions map[string]map[int]map[int]twoArgumentsFunctionType = map[string]map[int]map[int]twoArgumentsFunctionType{
	yastExpressionEqual: {
		DataTypeString: {
			DataTypeString: makeFunctionStringEqual}},
	yastExpressionContains: {
		DataTypeString: {
			DataTypeString: makeFunctionStringContains},
		DataTypeNetwork: {
			DataTypeAddress: makeFunctionNetworkContainsAddress},
		DataTypeSetOfStrings: {
			DataTypeString: makeFunctionSetOfStringContains},
		DataTypeSetOfNetworks: {
			DataTypeAddress: makeFunctionSetOfNetworksContainsAddress},
		DataTypeSetOfDomains: {
			DataTypeDomain: makeFunctionSetOfDomainsContains}}}

func (ctx *YastCtx) getAdjustedArgument(v interface{}, val ExpressionType, attr *AttributeDesignatorType) (ExpressionType, *AttributeDesignatorType, error) {
	a, err := ctx.unmarshalExpression(v)
	if err != nil {
		return nil, nil, err
	}

	switch f := a.(type) {
	case AttributeValueType:
		if val == nil {
			return &f, attr, nil
		}

		return nil, nil, ctx.errorf("Expected one immediate value and one attribute got both immediate values")

	case *SelectorType:
		if val == nil {
			return f, attr, nil
		}

		return nil, nil, ctx.errorf("Expected one immediate value and one attribute got both immediate values")

	case AttributeDesignatorType:
		if attr == nil {
			return val, &f, nil
		}

		return nil, nil, ctx.errorf("Expected one immediate value and one attribute got both attributes")
	}

	return nil, nil, ctx.errorf("Expected one immediate value and one attribute got %T", a)
}

func (ctx *YastCtx) getAdjustedArgumentPair(items interface{}) (ExpressionType, AttributeDesignatorType, error) {
	args, err := ctx.validateList(items, "target function arguments")
	if len(args) != 2 {
		return AttributeValueType{}, AttributeDesignatorType{},
			ctx.errorf("Expected 2 arguments got %d", len(args))
	}

	first, second, err := ctx.getAdjustedArgument(args[0], nil, nil)
	if err != nil {
		return AttributeValueType{}, AttributeDesignatorType{}, err
	}

	first, second, err = ctx.getAdjustedArgument(args[1], first, second)
	if err != nil {
		return AttributeValueType{}, AttributeDesignatorType{}, err
	}

	return first, *second, nil
}

func (ctx *YastCtx) unmarshalTargetMatchExpression(item interface{}) (MatchType, error) {
	e, err := ctx.validateMap(item, "target match expression")
	if err != nil {
		return MatchType{}, err
	}

	k, v, err := ctx.getSingleMapPair(e, "expression map")
	if err != nil {
		return MatchType{}, err
	}

	ID, err := ctx.validateString(k, "target function identifier")
	if err != nil {
		return MatchType{}, err
	}

	ctx.pushNodeSpec("%#v", ID)
	defer ctx.popNodeSpec()

	typeFunctionMap, ok := targetCompatibleExpressions[strings.ToLower(ID)]
	if !ok {
		return MatchType{}, fmt.Errorf("Unknown match function %s", ID)
	}

	first, second, err := ctx.getAdjustedArgumentPair(v)
	if err != nil {
		return MatchType{}, err
	}

	firstType := first.getResultType()
	secondType := second.getResultType()

	subTypeFunctionMap, ok := typeFunctionMap[firstType]
	if !ok {
		return MatchType{}, ctx.errorf("No function %s for arguments %s and %s", ID, DataTypeNames[firstType], DataTypeNames[secondType])
	}

	maker, ok := subTypeFunctionMap[secondType]
	if !ok {
		return MatchType{}, ctx.errorf("No function %s for arguments %s and %s", ID, DataTypeNames[firstType], DataTypeNames[secondType])
	}

	return MatchType{maker(first, second)}, nil
}

func (ctx *YastCtx) unmarshalTargetAllOfItem(v interface{}, i int, matches []MatchType) ([]MatchType, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	m, err := ctx.unmarshalTargetMatchExpression(v)
	if err != nil {
		return nil, err
	}

	return append(matches, m), nil
}

func (ctx *YastCtx) unmarshalTargetAllOf(item interface{}) (AllOfType, error) {
	e, err := ctx.validateMap(item, "target expression")
	if err != nil {
		return AllOfType{}, err
	}

	k, v, err := ctx.getSingleMapPair(e, "expression map")
	if err != nil {
		return AllOfType{}, err
	}

	ID, err := ctx.validateString(k, "target function identifier")
	if err != nil {
		return AllOfType{}, err
	}

	if ID == "all" {
		ctx.pushNodeSpec("%#v", ID)
		defer ctx.popNodeSpec()

		args, err := ctx.validateList(v, "list of target all expressions")
		if err != nil {
			return AllOfType{}, err
		}

		r := AllOfType{}
		r.Matches = make([]MatchType, 0)
		for i, arg := range args {
			r.Matches, err = ctx.unmarshalTargetAllOfItem(arg, i, r.Matches)
			if err != nil {
				return AllOfType{}, err
			}
		}

		return r, nil
	}

	m, err := ctx.unmarshalTargetMatchExpression(item)
	if err != nil {
		return AllOfType{}, err
	}

	return AllOfType{[]MatchType{m}}, nil
}

func (ctx *YastCtx) unmarshalTargetAnyOfItem(v interface{}, i int, allOfs []AllOfType) ([]AllOfType, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	a, err := ctx.unmarshalTargetAllOf(v)
	if err != nil {
		return nil, err
	}

	return append(allOfs, a), nil
}

func (ctx *YastCtx) unmarshalTargetAnyOf(item interface{}) (AnyOfType, error) {
	e, err := ctx.validateMap(item, "target expression")
	if err != nil {
		return AnyOfType{}, err
	}

	k, v, err := ctx.getSingleMapPair(e, "expression map")
	if err != nil {
		return AnyOfType{}, err
	}

	ID, err := ctx.validateString(k, "target function identifier")
	if err != nil {
		return AnyOfType{}, err
	}

	if ID == "any" {
		ctx.pushNodeSpec("%#v", ID)
		defer ctx.popNodeSpec()

		args, err := ctx.validateList(v, "list of target any expressions")
		if err != nil {
			return AnyOfType{}, err
		}

		r := AnyOfType{}
		r.AllOf = make([]AllOfType, 0)
		for i, arg := range args {
			r.AllOf, err = ctx.unmarshalTargetAnyOfItem(arg, i, r.AllOf)
			if err != nil {
				return AnyOfType{}, err
			}
		}

		return r, nil
	}

	m, err := ctx.unmarshalTargetMatchExpression(item)
	if err != nil {
		return AnyOfType{}, err
	}

	return AnyOfType{[]AllOfType{{[]MatchType{m}}}}, nil
}

func (ctx *YastCtx) unmarshalTargetItem(v interface{}, i int, anyOfs []AnyOfType) ([]AnyOfType, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	a, err := ctx.unmarshalTargetAnyOf(v)
	if err != nil {
		return nil, err
	}

	return append(anyOfs, a), nil
}

func (ctx *YastCtx) unmarshalTarget(m map[interface{}]interface{}) (TargetType, error) {
	tree, ok := m[yastTagTarget]
	if !ok {
		return TargetType{}, nil
	}

	ctx.pushNodeSpec(yastTagTarget)
	defer ctx.popNodeSpec()

	items, err := ctx.validateList(tree, "target")
	if err != nil {
		return TargetType{}, err
	}

	t := TargetType{}
	t.AnyOf = make([]AnyOfType, 0)

	for i, item := range items {
		t.AnyOf, err = ctx.unmarshalTargetItem(item, i, t.AnyOf)
		if err != nil {
			return TargetType{}, err
		}
	}

	return t, nil
}
