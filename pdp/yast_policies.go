package pdp

import "strings"

type paramUnmarshalerRCAType func (ctx *yastCtx, p *PolicyType, m map[interface{}]interface{}) error
type paramUnmarshalerPCAType func (ctx *yastCtx, p *PolicySetType, m map[interface{}]interface{}) error

var paramUnmarshalersRCA = map[string]paramUnmarshalerRCAType{yastTagMapperAlg: unmarshalMapperRCAParams}
var paramUnmarshalersPCA = map[string]paramUnmarshalerPCAType{yastTagMapperAlg: unmarshalMapperPCAParams}

func unmarshalMapperRCAParams(ctx *yastCtx, p *PolicyType, m map[interface{}]interface{}) error {
	v, ok  := m[yastTagMap]
	if !ok {
		return ctx.errorf("Missing map")
	}

	expr, err := ctx.unmarshalExpression(v)
	if err != nil {
		return err
	}

	exprType := expr.getResultType()
	if exprType != DataTypeString {
		return ctx.errorf("Expected %q expression but got %q", DataTypeNames[DataTypeString], DataTypeNames[exprType])
	}

	p.argument = expr

	p.rulesMap = make(map[string]*RuleType)
	for _, r := range p.Rules {
		p.rulesMap[r.ID] = &r
	}

	var defRule *RuleType = nil
	v, ok = m[yastTagDefault]
	if ok {
		ID, err := ctx.validateString(v, "default rule ID")
		if err != nil {
			return err
		}

		defRule, ok = p.rulesMap[ID]
		if !ok {
			return ctx.errorf("Unknown rule ID %s (expected rule defined in current policy)", ID)
		}
	}
	p.defaultRule = defRule

	var errRule *RuleType = nil
	v, ok = m[yastTagError]
	if ok {
		ID, err := ctx.validateString(v, "error rule ID")
		if err != nil {
			return err
		}

		errRule, ok = p.rulesMap[ID]
		if !ok {
			return ctx.errorf("Unknown rule ID %s (expected rule defined in current policy)", ID)
		}
	}
	p.errorRule = errRule

	return nil
}

func (ctx *yastCtx) unmarshalPolicy(m map[interface{}]interface{}, items interface{}) (PolicyType, error) {
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

	algID, alg, params, err := ctx.unmarshalRuleCombiningAlg(m)
	if err != nil {
		return pol, err
	}

	pol.ID = ID
	pol.Target = t
	pol.Rules = r
	pol.RuleCombiningAlg = alg

	paramsUnmarshaler, ok := paramUnmarshalersRCA[algID]
	if ok {
		if params == nil {
			return pol, ctx.errorf("Missing parameters for %q rule combining algorithm", algID)
		}

		err = paramsUnmarshaler(ctx, &pol, params)
		if err != nil {
			return pol, err
		}
	}

	return pol, nil
}

func (ctx *yastCtx) unmarshalPoliciesItem(v interface{}, i int, items []EvaluableType) ([]EvaluableType, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	p, err := ctx.unmarshalItem(v)
	if err != nil {
		return nil, err
	}

	return append(items, p), nil
}

func (ctx *yastCtx) unmarshalPolicies(item interface{}) ([]EvaluableType, error) {
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

func (ctx *yastCtx) unmarshalPolicyCombiningAlg(m map[interface{}]interface{}) (string, PolicyCombiningAlgType, map[interface{}]interface{}, error) {
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

func unmarshalMapperPCAParams(ctx *yastCtx, p *PolicySetType, m map[interface{}]interface{}) error {
	v, ok  := m[yastTagMap]
	if !ok {
		return ctx.errorf("Missing map")
	}

	expr, err := ctx.unmarshalExpression(v)
	if err != nil {
		return err
	}

	exprType := expr.getResultType()
	if exprType != DataTypeString {
		return ctx.errorf("Expected %q expression but got %q", DataTypeNames[DataTypeString], DataTypeNames[exprType])
	}

	p.argument = expr

	p.policiesMap = make(map[string]EvaluableType)
	for _, e := range p.Policies {
		p.policiesMap[e.getID()] = e
	}

	var defPolicy EvaluableType = nil
	v, ok = m[yastTagDefault]
	if ok {
		ID, err := ctx.validateString(v, "default policy or set ID")
		if err != nil {
			return err
		}

		defPolicy, ok = p.policiesMap[ID]
		if !ok {
			return ctx.errorf("Unknown policy or set ID %s (expected item defined in current set)", ID)
		}
	}
	p.defaultPolicy = defPolicy

	var errPolicy EvaluableType = nil
	v, ok = m[yastTagError]
	if ok {
		ID, err := ctx.validateString(v, "error policy or set ID")
		if err != nil {
			return err
		}

		errPolicy, ok = p.policiesMap[ID]
		if !ok {
			return ctx.errorf("Unknown policy or set ID %s (expected item defined in current set)", ID)
		}
	}
	p.errorPolicy = errPolicy

	return nil
}

func (ctx *yastCtx) unmarshalPolicySet(m map[interface{}]interface{}, items interface{}) (PolicySetType, error) {
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

	algID, alg, params, err := ctx.unmarshalPolicyCombiningAlg(m)
	if err != nil {
		return pol, err
	}

	pol.ID = ID
	pol.Target = t
	pol.Policies = p
	pol.PolicyCombiningAlg = alg

	paramsUnmarshaler, ok := paramUnmarshalersPCA[algID]
	if ok {
		if params == nil {
			return pol, ctx.errorf("Missing parameters for %q policies combining algorithm", algID)
		}

		err = paramsUnmarshaler(ctx, &pol, params)
		if err != nil {
			return pol, err
		}
	}

	return pol, nil
}

func (ctx *yastCtx) unmarshalItem(v interface{}) (EvaluableType, error) {
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
