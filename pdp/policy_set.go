package pdp

import "fmt"

type PolicyCombiningAlgType func(policySet *PolicySetType, ctx *Context) ResponseType

var PolicyCombiningAlgMap map[string]PolicyCombiningAlgType = map[string]PolicyCombiningAlgType{
	yastTagFirstApplicableEffectAlg: FirstApplicableEffectPCA,
	yastTagDenyOverridesAlg:         DenyOverridesPCA,
	yastTagMapperAlg:                MapperPCA}

type PolicySetType struct {
	ID                 string
	Target             TargetType
	Policies           []EvaluableType
	Obligations        []AttributeAssignmentExpressionType
	PolicyCombiningAlg PolicyCombiningAlgType

	argument      ExpressionType
	policiesMap   map[string]EvaluableType
	defaultPolicy EvaluableType
	errorPolicy   EvaluableType
}

func (p PolicySetType) getID() string {
	return p.ID
}

func (p PolicySetType) Calculate(ctx *Context) ResponseType {
	match, err := p.Target.calculate(ctx)
	if err != nil {
		return combineEffectAndStatus(err, p.ID, p.PolicyCombiningAlg(&p, ctx))
	}

	if !match {
		return ResponseType{EffectNotApplicable, "Ok", nil}
	}

	r := p.PolicyCombiningAlg(&p, ctx)
	if r.Effect == EffectDeny || r.Effect == EffectPermit {
		r.Obligations = append(r.Obligations, p.Obligations...)
	}

	return r
}

func FirstApplicableEffectPCA(policySet *PolicySetType, ctx *Context) ResponseType {
	for _, p := range policySet.Policies {
		r := p.Calculate(ctx)
		if r.Effect != EffectNotApplicable {
			return r
		}
	}

	return ResponseType{EffectNotApplicable, "Ok", nil}
}

func DenyOverridesPCA(policySet *PolicySetType, ctx *Context) ResponseType {
	status := ""
	obligations := make([]AttributeAssignmentExpressionType, 0)

	indetD := 0
	indetP := 0
	indetDP := 0

	permits := 0

	for _, p := range policySet.Policies {
		r := p.Calculate(ctx)
		if r.Effect == EffectDeny {
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

func MapperPCA(policySet *PolicySetType, ctx *Context) ResponseType {
	v, err := policySet.argument.calculate(ctx)
	if err != nil {
		if policySet.errorPolicy != nil {
			return policySet.errorPolicy.Calculate(ctx)
		}

		return ResponseType{EffectIndeterminate, fmt.Sprintf("Mapper Policy Combining Algorithm: %s", err), nil}
	}

	ID, err := ExtractStringValue(v, "argument")
	if err != nil {
		if policySet.errorPolicy != nil {
			return policySet.errorPolicy.Calculate(ctx)
		}

		return ResponseType{EffectIndeterminate, fmt.Sprintf("Mapper Policy Combining Algorithm: %s", err), nil}
	}

	policy, ok := policySet.policiesMap[ID]
	if ok {
		return policy.Calculate(ctx)
	}

	if policySet.defaultPolicy != nil {
		return policySet.defaultPolicy.Calculate(ctx)
	}

	return ResponseType{EffectNotApplicable, "Ok", nil}
}
