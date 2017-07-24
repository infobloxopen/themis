package pdp

type mapperPCA struct {
	argument  expression
	policies  map[string]Evaluable
	def       Evaluable
	err       Evaluable
	algorithm policyCombiningAlg
}

func collectSubPolicies(IDs []string, m map[string]Evaluable) []Evaluable {
	policies := []Evaluable{}
	for _, ID := range IDs {
		policy, ok := m[ID]
		if ok {
			policies = append(policies, policy)
		}
	}

	return policies
}

func (a mapperPCA) describe() string {
	return "mapper"
}

func (a mapperPCA) calculateErrorPolicy(ctx *Context, err error) Response {
	if a.err != nil {
		return a.err.Calculate(ctx)
	}

	return Response{EffectIndeterminate, bindError(err, a.describe()), nil}
}

func (a mapperPCA) getPoliciesMap(policies []Evaluable) map[string]Evaluable {
	if a.policies != nil {
		return a.policies
	}

	m := make(map[string]Evaluable)
	for _, policy := range policies {
		m[policy.GetID()] = policy
	}

	return m
}

func (a mapperPCA) execute(policies []Evaluable, ctx *Context) Response {
	v, err := a.argument.calculate(ctx)
	if err != nil {
		switch err.(type) {
		case *missingValueError:
			if a.def != nil {
				return a.def.Calculate(ctx)
			}
		}

		return a.calculateErrorPolicy(ctx, err)
	}

	if a.algorithm != nil {
		IDs, err := getSetOfIDs(v)
		if err != nil {
			return a.calculateErrorPolicy(ctx, err)
		}

		r := a.algorithm.execute(collectSubPolicies(IDs, a.getPoliciesMap(policies)), ctx)
		if r.Effect == EffectNotApplicable && a.def != nil {
			return a.def.Calculate(ctx)
		}

		return r
	}

	ID, err := v.str()
	if err != nil {
		return a.calculateErrorPolicy(ctx, err)
	}

	if a.policies != nil {
		policy, ok := a.policies[ID]
		if ok {
			return policy.Calculate(ctx)
		}
	} else {
		for _, policy := range policies {
			if policy.GetID() == ID {
				return policy.Calculate(ctx)
			}
		}
	}

	if a.def != nil {
		return a.def.Calculate(ctx)
	}

	return Response{EffectNotApplicable, nil, nil}
}
