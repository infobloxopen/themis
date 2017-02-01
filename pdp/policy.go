package pdp

import "fmt"

type RuleCombiningAlgType func(policy *PolicyType, ctx *Context) ResponseType

var RuleCombiningAlgMap map[string]RuleCombiningAlgType = map[string]RuleCombiningAlgType{
	yastTagFirstApplicableEffectAlg: FirstApplicableEffectRCA,
	yastTagDenyOverridesAlg:         DenyOverridesRCA,
	yastTagMapperAlg:                MapperRCA}

type PolicyType struct {
	ID               string
	Target           TargetType
	Rules            []RuleType
	Obligations      []AttributeAssignmentExpressionType
	RuleCombiningAlg RuleCombiningAlgType

	argument     ExpressionType
	rulesMap     map[string]*RuleType
	defaultRule  *RuleType
	errorRule    *RuleType
}

func (p PolicyType) getID() string {
	return p.ID
}

func (p PolicyType) Calculate(ctx *Context) ResponseType {
	match, err := p.Target.calculate(ctx)
	if err != nil {
		return combineEffectAndStatus(err, p.ID, p.RuleCombiningAlg(&p, ctx))
	}

	if !match {
		return ResponseType{EffectNotApplicable, "Ok", nil}
	}

	r := p.RuleCombiningAlg(&p, ctx)
	if r.Effect == EffectDeny || r.Effect == EffectPermit {
		r.Obligations = append(r.Obligations, p.Obligations...)
	}

	return r
}

func FirstApplicableEffectRCA(policy *PolicyType, ctx *Context) ResponseType {
	for _, rule := range policy.Rules {
		r := rule.calculate(ctx)
		if r.Effect != EffectNotApplicable {
			return r
		}
	}

	return ResponseType{EffectNotApplicable, "Ok", nil}
}

func DenyOverridesRCA(policy *PolicyType, ctx *Context) ResponseType {
	status := ""
	obligations := make([]AttributeAssignmentExpressionType, 0)

	indetD := 0
	indetP := 0
	indetDP := 0

	permits := 0

	for _, rule := range policy.Rules {
		r := rule.calculate(ctx)
		if r.Effect == EffectDeny {
			obligations = append(obligations, r.Obligations...)
			return r
		}

		if r.Effect == EffectPermit {
			permits += 1
			obligations = append(obligations, r.Obligations...)
			continue
		}

		if r.Effect == EffectNotApplicable {
			continue
		}

		if r.Effect == EffectIndeterminateD {
			indetD += 1
		} else {
			if r.Effect == EffectIndeterminateP {
				indetP += 1
			} else {
				indetDP += 1
			}

		}

		if len(status) > 0 {
			status += ", "
		}

		status += r.Status
	}

	if indetDP > 0 || (indetD > 0 && (indetP > 0 || permits > 0)) {
		return ResponseType{EffectIndeterminateDP, status, nil}
	}

	if indetD > 0 {
		return ResponseType{EffectIndeterminateD, status, nil}
	}

	if permits > 0 {
		return ResponseType{EffectPermit, "Ok", obligations}
	}

	if indetP > 0 {
		return ResponseType{EffectIndeterminateP, status, nil}
	}

	return ResponseType{EffectNotApplicable, "Ok", nil}
}

func MapperRCA(policy *PolicyType, ctx *Context) ResponseType {
	v, err := policy.argument.calculate(ctx)
	if err != nil {
		if policy.errorRule != nil {
			return policy.errorRule.calculate(ctx)
		}

		return ResponseType{EffectIndeterminate, fmt.Sprintf("Mapper Rule Combining Algorithm: %s", err), nil}
	}

	ID, err := ExtractStringValue(v, "argument")
	if err != nil {
		if policy.errorRule != nil {
			return policy.errorRule.calculate(ctx)
		}

		return ResponseType{EffectIndeterminate, fmt.Sprintf("Mapper Rule Combining Algorithm: %s", err), nil}
	}

	rule, ok := policy.rulesMap[ID]
	if ok {
		return rule.calculate(ctx)
	}

	if policy.defaultRule != nil {
		return policy.defaultRule.calculate(ctx)
	}

	return ResponseType{EffectNotApplicable, "Ok", nil}
}
