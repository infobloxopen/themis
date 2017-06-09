package pdp

import (
	"fmt"
	"strings"
)

type paramUnmarshalerRCAType func(ctx *YastCtx, p *PolicyType, root bool, m map[interface{}]interface{}) (interface{}, error)
type paramUnmarshalerPCAType func(ctx *YastCtx, p *PolicySetType, root bool, m map[interface{}]interface{}) (interface{}, error)

var paramUnmarshalersRCA = map[string]paramUnmarshalerRCAType{}
var paramUnmarshalersPCA = map[string]paramUnmarshalerPCAType{}

func init() {
	paramUnmarshalersRCA[yastTagMapperAlg] = unmarshalMapperRCAParams
	paramUnmarshalersPCA[yastTagMapperAlg] = unmarshalMapperPCAParams
}

func (ctx *YastCtx) findSpecialRule(m map[interface{}]interface{}, tag string, rules map[string]*RuleType) (*RuleType, error) {
	v, ok := m[tag]
	if !ok {
		return nil, nil
	}

	ID, err := ctx.validateString(v, fmt.Sprintf("%s rule ID", tag))
	if err != nil {
		return nil, err
	}

	rule, ok := rules[ID]
	if !ok {
		return nil, ctx.errorf("Unknown rule ID %s (expected rule defined in current policy)", ID)
	}

	return rule, nil
}

func (ctx *YastCtx) unmarshalRuleCombiningAlg(p *PolicyType, root bool, m map[interface{}]interface{}) (RuleCombiningAlgType, interface{}, error) {
	algID, alg, params, err := ctx.extractRuleCombiningAlg(m)
	if err != nil {
		return nil, nil, err
	}

	paramsUnmarshaler, ok := paramUnmarshalersRCA[algID]
	if !ok {
		return alg, nil, nil
	}

	ctx.pushNodeSpec("%q", algID)
	defer ctx.popNodeSpec()

	if params == nil {
		return alg, nil, ctx.errorf("Missing parameters for %q rule combining algorithm", algID)
	}

	algParams, err := paramsUnmarshaler(ctx, p, root, params)
	if err != nil {
		return alg, nil, err
	}

	return alg, algParams, nil
}

func unmarshalMapperRCAParams(ctx *YastCtx, p *PolicyType, root bool, m map[interface{}]interface{}) (interface{}, error) {
	v, ok := m[yastTagMap]
	if !ok {
		return nil, ctx.errorf("Missing map")
	}

	expr, err := ctx.unmarshalExpression(v)
	if err != nil {
		return nil, err
	}

	exprType := expr.getResultType()

	params := MapperRCAParams{Argument: expr}

	rulesMap := make(map[string]*RuleType)
	for _, r := range p.Rules {
		tmp := r
		rulesMap[r.ID] = &tmp
	}

	if root {
		params.RulesMap = rulesMap
	}

	params.DefaultRule, err = ctx.findSpecialRule(m, yastTagDefault, rulesMap)
	if err != nil {
		return params, err
	}

	params.ErrorRule, err = ctx.findSpecialRule(m, yastTagError, rulesMap)
	if err != nil {
		return params, err
	}

	_, ok = m[yastTagAlg]
	if ok {
		subAlg, subAlgParams, err := ctx.unmarshalRuleCombiningAlg(p, false, m)
		if err != nil {
			return params, err
		}

		params.SubAlg = subAlg
		params.AlgParams = subAlgParams

		if exprType != DataTypeString && exprType != DataTypeSetOfStrings {
			return params, ctx.errorf("Expected %q or %q expression but got %q", DataTypeNames[DataTypeString], DataTypeNames[DataTypeSetOfStrings], DataTypeNames[exprType])
		}
	} else {
		if exprType != DataTypeString {
			return params, ctx.errorf("Expected %q expression but got %q", DataTypeNames[DataTypeString], DataTypeNames[exprType])
		}
	}

	return params, nil
}

func (ctx *YastCtx) unmarshalPolicy(m map[interface{}]interface{}, items interface{}) (PolicyType, error) {
	pol := PolicyType{}

	ID, err := ctx.extractString(m, yastTagID, "policy id")
	if err != nil {
		return pol, err
	}

	ctx.pushNodeSpec("%#v", ID)
	defer ctx.popNodeSpec()

	t, err := ctx.unmarshalTarget(m)
	if err != nil {
		return pol, err
	}

	r, err := ctx.unmarshalRules(items)
	if err != nil {
		return pol, err
	}

	pol.ID = ID
	pol.Target = t
	pol.Rules = r

	alg, params, err := ctx.unmarshalRuleCombiningAlg(&pol, true, m)
	if err != nil {
		return pol, err
	}

	pol.RuleCombiningAlg = alg
	pol.AlgParams = params

	return pol, nil
}

func (ctx *YastCtx) unmarshalPoliciesItem(v interface{}, i int, items []EvaluableType) ([]EvaluableType, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	p, err := ctx.unmarshalItem(v)
	if err != nil {
		return nil, err
	}

	return append(items, p), nil
}

func (ctx *YastCtx) unmarshalPolicies(item interface{}) ([]EvaluableType, error) {
	ctx.pushNodeSpec(yastTagPolicies)
	defer ctx.popNodeSpec()

	items, err := ctx.validateList(item, "list of policies")
	if err != nil {
		return nil, err
	}

	r := make([]EvaluableType, 0)
	for i, p := range items {
		r, err = ctx.unmarshalPoliciesItem(p, i, r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

func (ctx *YastCtx) extractPolicyCombiningAlg(m map[interface{}]interface{}) (string, PolicyCombiningAlgType, map[interface{}]interface{}, error) {
	s, algMap, err := ctx.extractStringOrMapDef(m, yastTagAlg, yastTagDefaultAlg, nil, "policy combining algorithm")
	if err != nil {
		return "", nil, nil, err
	}

	if algMap != nil {
		s, err = ctx.extractString(algMap, yastTagID, "policy combining algorithm id")
		if err != nil {
			return "", nil, nil, err
		}
	}

	ID := strings.ToLower(s)
	a, ok := PolicyCombiningAlgMap[ID]
	if !ok {
		return "", nil, nil, ctx.errorf("Excpected policy combining algorithm but got %s", s)
	}

	return ID, a, algMap, nil
}

func (ctx *YastCtx) findSpecialPolicy(m map[interface{}]interface{}, tag string, policies map[string]EvaluableType) (EvaluableType, error) {
	v, ok := m[tag]
	if !ok {
		return nil, nil
	}

	ID, err := ctx.validateString(v, fmt.Sprintf("%s policy or set ID", tag))
	if err != nil {
		return nil, err
	}

	policy, ok := policies[ID]
	if !ok {
		return nil, ctx.errorf("Unknown policy or set ID %s (expected item defined in current set)", ID)
	}

	return policy, nil
}

func (ctx *YastCtx) unmarshalPolicyCombiningAlg(p *PolicySetType, root bool, m map[interface{}]interface{}) (PolicyCombiningAlgType, interface{}, error) {
	algID, alg, params, err := ctx.extractPolicyCombiningAlg(m)
	if err != nil {
		return nil, nil, err
	}

	paramsUnmarshaler, ok := paramUnmarshalersPCA[algID]
	if !ok {
		return alg, nil, nil
	}

	ctx.pushNodeSpec("%q", algID)
	defer ctx.popNodeSpec()

	if params == nil {
		return alg, nil, ctx.errorf("Missing parameters for %q policies combining algorithm", algID)
	}

	algParams, err := paramsUnmarshaler(ctx, p, root, params)
	if err != nil {
		return alg, nil, err
	}

	return alg, algParams, nil
}

func unmarshalMapperPCAParams(ctx *YastCtx, p *PolicySetType, root bool, m map[interface{}]interface{}) (interface{}, error) {
	v, ok := m[yastTagMap]
	if !ok {
		return nil, ctx.errorf("Missing map")
	}

	expr, err := ctx.unmarshalExpression(v)
	if err != nil {
		return nil, err
	}

	exprType := expr.getResultType()

	params := MapperPCAParams{Argument: expr}

	policiesMap := make(map[string]EvaluableType)
	for _, e := range p.Policies {
		policiesMap[e.GetID()] = e
	}

	if root {
		params.PoliciesMap = policiesMap
	}

	params.DefaultPolicy, err = ctx.findSpecialPolicy(m, yastTagDefault, policiesMap)
	if err != nil {
		return nil, err
	}

	params.ErrorPolicy, err = ctx.findSpecialPolicy(m, yastTagError, policiesMap)
	if err != nil {
		return nil, err
	}

	_, ok = m[yastTagAlg]
	if ok {
		subAlg, subAlgParams, err := ctx.unmarshalPolicyCombiningAlg(p, false, m)
		if err != nil {
			return params, nil
		}

		params.SubAlg = subAlg
		params.AlgParams = subAlgParams

		if exprType != DataTypeString && exprType != DataTypeSetOfStrings {
			return nil, ctx.errorf("Expected %q or %q expression but got %q", DataTypeNames[DataTypeString], DataTypeNames[DataTypeSetOfStrings], DataTypeNames[exprType])
		}
	} else {
		if exprType != DataTypeString {
			return nil, ctx.errorf("Expected %q expression but got %q", DataTypeNames[DataTypeString], DataTypeNames[exprType])
		}
	}

	return params, nil
}

func (ctx *YastCtx) unmarshalPolicySet(m map[interface{}]interface{}, items interface{}) (PolicySetType, error) {
	pol := PolicySetType{}

	ID, err := ctx.extractString(m, yastTagID, "policy set id")
	if err != nil {
		return pol, err
	}

	ctx.pushNodeSpec("%#v", ID)
	defer ctx.popNodeSpec()

	t, err := ctx.unmarshalTarget(m)
	if err != nil {
		return pol, err
	}

	p, err := ctx.unmarshalPolicies(items)
	if err != nil {
		return pol, err
	}

	pol.ID = ID
	pol.Target = t
	pol.Policies = p

	alg, algParams, err := ctx.unmarshalPolicyCombiningAlg(&pol, true, m)
	if err != nil {
		return pol, err
	}

	pol.PolicyCombiningAlg = alg
	pol.AlgParams = algParams

	return pol, nil
}

func (ctx *YastCtx) unmarshalItem(v interface{}) (EvaluableType, error) {
	r, err := ctx.validateMap(v, "policy or policy set")
	if err != nil {
		return nil, err
	}

	rules, rules_ok := r[yastTagRules]
	policies, policies_ok := r[yastTagPolicies]

	if rules_ok && policies_ok {
		return nil, ctx.errorf("Expected rules (for policy) or policies (for policy set) but got both")
	}

	if rules_ok {
		return ctx.unmarshalPolicy(r, rules)
	}

	if policies_ok {
		return ctx.unmarshalPolicySet(r, policies)
	}

	return nil, ctx.errorf("Expected rules (for policy) or policies (for policy set) but got nothing")
}
