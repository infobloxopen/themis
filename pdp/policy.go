package pdp

import "fmt"

type RuleCombiningAlg interface {
	execute(rules []*Rule, ctx *Context) Response
}

type RuleCombiningAlgMaker func(rules []*Rule, params interface{}) RuleCombiningAlg

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
	algorithm   RuleCombiningAlg
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

func (p *Policy) Append(path []string, v interface{}) (Evaluable, error) {
	if p.hidden {
		return p, newHiddenPolicyModificationError()
	}

	if len(path) > 0 {
		return p, bindError(newTooLongPathPolicyModificationError(path), p.id)
	}

	child, ok := v.(*Rule)
	if !ok {
		return p, bindError(newInvalidPolicyItemTypeError(v), p.id)
	}

	_, ok = child.GetID()
	if !ok {
		return p, bindError(newHiddenRuleAppendError(), p.id)
	}

	return p.putChild(child), nil
}

func (p *Policy) Delete(path []string) (Evaluable, error) {
	if p.hidden {
		return p, newHiddenPolicyModificationError()
	}

	if len(path) <= 0 {
		return p, bindError(newTooShortPathPolicyModificationError(), p.id)
	}

	ID := path[0]

	if len(path) > 1 {
		return p, bindError(newTooLongPathPolicyModificationError(path[1:]), p.id)
	}

	r, err := p.delChild(ID)
	if err != nil {
		return p, bindError(err, p.id)
	}

	return r, nil
}

func (p *Policy) putChild(child *Rule) *Policy {
	ID, _ := child.GetID()
	for i, old := range p.rules {
		if rID, ok := old.GetID(); ok && rID == ID {
			rules := []*Rule{}
			if i > 0 {
				rules = append(rules, p.rules[:i]...)
			}

			rules = append(rules, child)

			if i+1 < len(p.rules) {
				rules = append(rules, p.rules[i+1:]...)
			}

			algorithm := p.algorithm
			if m, ok := algorithm.(mapperRCA); ok {
				algorithm = m.add(ID, child, old)
			}

			return &Policy{
				id:          p.id,
				target:      p.target,
				rules:       rules,
				obligations: p.obligations,
				algorithm:   algorithm}
		}
	}

	rules := p.rules
	if rules == nil {
		rules = []*Rule{child}
	} else {
		rules = append(rules, child)
	}

	algorithm := p.algorithm
	if m, ok := algorithm.(mapperRCA); ok {
		algorithm = m.add(ID, child, nil)
	}

	return &Policy{
		id:          p.id,
		target:      p.target,
		rules:       rules,
		obligations: p.obligations,
		algorithm:   algorithm}
}

func (p *Policy) delChild(ID string) (*Policy, error) {
	for i, old := range p.rules {
		if rID, ok := old.GetID(); ok && rID == ID {
			rules := []*Rule{}
			if i > 0 {
				rules = append(rules, p.rules[:i]...)
			}

			if i+1 < len(p.rules) {
				rules = append(rules, p.rules[i+1:]...)
			}

			algorithm := p.algorithm
			if m, ok := algorithm.(mapperRCA); ok {
				algorithm = m.del(ID, old)
			}

			return &Policy{
				id:          p.id,
				target:      p.target,
				rules:       rules,
				obligations: p.obligations,
				algorithm:   algorithm}, nil
		}
	}

	return nil, newMissingPolicyChildError(ID)
}

type firstApplicableEffectRCA struct {
}

func makeFirstApplicableEffectRCA(rules []*Rule, params interface{}) RuleCombiningAlg {
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

func makeDenyOverridesRCA(rules []*Rule, params interface{}) RuleCombiningAlg {
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
		err = bindError(newMultiError(errs), a.describe())
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
