package pdp

import "strings"

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

	alg, err := ctx.unmarshalRuleCombiningAlg(m)
	if err != nil {
		return pol, err
	}

	pol.ID = ID
	pol.Target = t
	pol.Rules = r
	pol.RuleCombiningAlg = alg

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

func (ctx *yastCtx) unmarshalPolicyCombiningAlg(m map[interface{}]interface{}) (PolicyCombiningAlgType, error) {
	s, err := ctx.extractStringDef(m, yastTagAlg, yastTagDefaultAlg, "policy combining algorithm")
	if err != nil {
		return nil, err
	}

	a, ok := PolicyCombiningAlgMap[strings.ToLower(s)]
	if !ok {
		return nil, ctx.errorf("Excpected policy combining algorithm but got %s", s)
	}

	return a, nil
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

	alg, err := ctx.unmarshalPolicyCombiningAlg(m)
	if err != nil {
		return pol, err
	}

	pol.ID = ID
	pol.Target = t
	pol.Policies = p
	pol.PolicyCombiningAlg = alg

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
