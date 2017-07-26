package pdp

import "fmt"

type ruleCombiningAlg interface {
	execute(rules []*Rule, ctx *Context) Response
}

type RuleCombiningAlgMaker func(rules []*Rule, params interface{}) ruleCombiningAlg

var (
	firstApplicableEffectRCAInstance = firstApplicableEffectRCA{}
	denyOverridesRCAInstance         = denyOverridesRCA{}

	RuleCombiningAlgs = map[string]RuleCombiningAlgMaker{
		"firstapplicableeffect": makeFirstApplicableEffectRCA,
		"denyoverrides":         makeDenyOverridesRCA}

	RuleCombiningParamAlgs = map[string]RuleCombiningAlgMaker{
		"mapper": makeMapperRCA}
)

type Policy struct {
	id          string
	hidden      bool
	target      Target
	rules       []*Rule
	obligations []AttributeAssignmentExpression
	algorithm   ruleCombiningAlg
}

func NewPolicy(ID string, hidden bool, target Target, rules []*Rule, makeRCA RuleCombiningAlgMaker, params interface{}, obligations []AttributeAssignmentExpression) *Policy {
	return &Policy{
		id:          ID,
		hidden:      hidden,
		target:      target,
		rules:       rules,
		obligations: obligations,
		algorithm:   makeRCA(rules, params)}
}

func (p *Policy) describe() string {
	if pid, ok := p.GetID(); ok {
		return fmt.Sprintf("policy %q", pid)
	}

	return "hidden policy"
}

func (p *Policy) GetID() (string, bool) {
	return p.id, !p.hidden
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

func makeFirstApplicableEffectRCA(rules []*Rule, params interface{}) ruleCombiningAlg {
	return firstApplicableEffectRCAInstance
}

func (a firstApplicableEffectRCA) execute(rules []*Rule, ctx *Context) Response {
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

func makeDenyOverridesRCA(rules []*Rule, params interface{}) ruleCombiningAlg {
	return denyOverridesRCAInstance
}

func (a denyOverridesRCA) describe() string {
	return "deny overrides"
}

func (a denyOverridesRCA) execute(rules []*Rule, ctx *Context) Response {
	errs := []error{}
	obligations := make([]AttributeAssignmentExpression, 0)

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
