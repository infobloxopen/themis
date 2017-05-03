package pdp

type PolicyCombiningAlgType func(policies []EvaluableType, params interface{}, ctx *Context) ResponseType

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
	AlgParams          interface{}
}

func (p PolicySetType) getID() string {
	return p.ID
}

func (p PolicySetType) Calculate(ctx *Context) ResponseType {
	match, err := p.Target.calculate(ctx)
	if err != nil {
		return combineEffectAndStatus(err, p.ID, p.PolicyCombiningAlg(p.Policies, p.AlgParams, ctx))
	}

	if !match {
		return ResponseType{EffectNotApplicable, "Ok", nil}
	}

	r := p.PolicyCombiningAlg(p.Policies, p.AlgParams, ctx)
	if r.Effect == EffectDeny || r.Effect == EffectPermit {
		r.Obligations = append(r.Obligations, p.Obligations...)
	}

	return r
}

func FirstApplicableEffectPCA(policies []EvaluableType, params interface{}, ctx *Context) ResponseType {
	for _, p := range policies {
		r := p.Calculate(ctx)
		if r.Effect != EffectNotApplicable {
			return r
		}
	}

	return ResponseType{EffectNotApplicable, "Ok", nil}
}

func DenyOverridesPCA(policies []EvaluableType, params interface{}, ctx *Context) ResponseType {
	status := ""
	obligations := make([]AttributeAssignmentExpressionType, 0)

	indetD := 0
	indetP := 0
	indetDP := 0

	permits := 0

	for _, p := range policies {
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
