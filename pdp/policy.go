package pdp

import "fmt"

type ruleCombiningAlg interface {
	execute(rules []*rule, ctx *Context) Response
}

type Policy struct {
	id          string
	target      target
	rules       []*rule
	obligations []attributeAssignmentExpression
	algorithm   ruleCombiningAlg
}

func (p *Policy) describe() string {
	return fmt.Sprintf("policy %q", p.id)
}

func (p *Policy) GetID() string {
	return p.id
}

func (p *Policy) Calculate(ctx *Context) Response {
	match, err := p.target.calculate(ctx)
	if err != nil {
		r := combineEffectAndStatus(err, p.algorithm.execute(p.rules, ctx))
		if r.status != nil {
			r.status = bindError(r.status, p.describe())
		}
		return r
	}

	if !match {
		return Response{EffectNotApplicable, nil, nil}
	}

	r := p.algorithm.execute(p.rules, ctx)
	if r.Effect == EffectDeny || r.Effect == EffectPermit {
		r.obligations = append(r.obligations, p.obligations...)
	}

	if r.status != nil {
		r.status = bindError(r.status, p.describe())
	}

	return r
}

type firstApplicableEffectRCA struct {
}

func (a firstApplicableEffectRCA) execute(rules []*rule, ctx *Context) Response {
	for _, rule := range rules {
		r := rule.calculate(ctx)
		if r.Effect != EffectNotApplicable {
			return r
		}
	}

	return Response{EffectNotApplicable, nil, nil}
}

type denyOverridesRCA struct {
}

func (a denyOverridesRCA) describe() string {
	return "deny overrides"
}

func (a denyOverridesRCA) execute(rules []*rule, ctx *Context) Response {
	errs := []error{}
	obligations := make([]attributeAssignmentExpression, 0)

	indetD := 0
	indetP := 0
	indetDP := 0

	permits := 0

	for _, rule := range rules {
		r := rule.calculate(ctx)
		if r.Effect == EffectDeny {
			obligations = append(obligations, r.obligations...)
			return r
		}

		if r.Effect == EffectPermit {
			permits++
			obligations = append(obligations, r.obligations...)
			continue
		}

		if r.Effect == EffectNotApplicable {
			continue
		}

		if r.Effect == EffectIndeterminateD {
			indetD++
		} else {
			if r.Effect == EffectIndeterminateP {
				indetP++
			} else {
				indetDP++
			}

		}

		errs = append(errs, r.status)
	}

	var err boundError
	if len(errs) > 1 {
		err = mewMultiError(errs, a.describe())
	} else if len(errs) > 0 {
		err = bindError(errs[0], a.describe())
	}

	if indetDP > 0 || (indetD > 0 && (indetP > 0 || permits > 0)) {
		return Response{EffectIndeterminateDP, err, nil}
	}

	if indetD > 0 {
		return Response{EffectIndeterminateD, err, nil}
	}

	if permits > 0 {
		return Response{EffectPermit, nil, obligations}
	}

	if indetP > 0 {
		return Response{EffectIndeterminateP, err, nil}
	}

	return Response{EffectNotApplicable, nil, nil}
}
