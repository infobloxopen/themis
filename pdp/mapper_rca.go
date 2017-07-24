package pdp

type mapperRCA struct {
	argument  expression
	rules     map[string]*rule
	def       *rule
	err       *rule
	algorithm ruleCombiningAlg
}

func getSetOfIDs(v attributeValue) ([]string, error) {
	ID, err := v.str()
	if err == nil {
		return []string{ID}, nil
	}

	setIDs, err := v.setOfStrings()
	if err == nil {
		return sortSetOfStrings(setIDs), nil
	}

	listIDs, err := v.listOfStrings()
	if err == nil {
		return listIDs, nil
	}

	return nil, newMapperArgumentTypeError(v.t)
}

func collectSubRules(IDs []string, m map[string]*rule) []*rule {
	rules := []*rule{}
	for _, ID := range IDs {
		rule, ok := m[ID]
		if ok {
			rules = append(rules, rule)
		}
	}

	return rules
}

func (a mapperRCA) describe() string {
	return "mapper"
}

func (a mapperRCA) calculateErrorRule(ctx *Context, err error) Response {
	if a.err != nil {
		return a.err.calculate(ctx)
	}

	return Response{EffectIndeterminate, bindError(err, a.describe()), nil}
}

func (a mapperRCA) getRulesMap(rules []*rule) map[string]*rule {
	if a.rules != nil {
		return a.rules
	}

	m := make(map[string]*rule)
	for _, rule := range rules {
		m[rule.id] = rule
	}

	return m
}

func (a mapperRCA) execute(rules []*rule, ctx *Context) Response {
	v, err := a.argument.calculate(ctx)
	if err != nil {
		switch err.(type) {
		case *missingValueError:
			if a.def != nil {
				return a.def.calculate(ctx)
			}
		}

		return a.calculateErrorRule(ctx, err)
	}

	if a.algorithm != nil {
		IDs, err := getSetOfIDs(v)
		if err != nil {
			return a.calculateErrorRule(ctx, err)
		}

		r := a.algorithm.execute(collectSubRules(IDs, a.getRulesMap(rules)), ctx)
		if r.Effect == EffectNotApplicable && a.def != nil {
			return a.def.calculate(ctx)
		}

		return r
	}

	ID, err := v.str()
	if err != nil {
		return a.calculateErrorRule(ctx, err)
	}

	if a.rules != nil {
		rule, ok := a.rules[ID]
		if ok {
			return rule.calculate(ctx)
		}
	} else {
		for _, rule := range rules {
			if rule.id == ID {
				return rule.calculate(ctx)
			}
		}
	}

	if a.def != nil {
		return a.def.calculate(ctx)
	}

	return Response{EffectNotApplicable, nil, nil}
}
