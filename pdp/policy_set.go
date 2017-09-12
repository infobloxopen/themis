package pdp

import "fmt"

// PolicyCombiningAlg represent abstract policy combining algorithm.
// The algorithm defines how to evaluate child policy sets and policies
// for given policy and how to get paticular result.
type PolicyCombiningAlg interface {
	execute(rules []Evaluable, ctx *Context) Response
}

// PolicyCombiningAlgMaker creates instance of policy combining algorithm.
// The function accepts set of child policy sets and policies and parameters
// of algorithm.
type PolicyCombiningAlgMaker func(policies []Evaluable, params interface{}) PolicyCombiningAlg

var (
	firstApplicableEffectPCAInstance = firstApplicableEffectPCA{}
	denyOverridesPCAInstance         = denyOverridesPCA{}

	// PolicyCombiningAlgs defines map of algorithm id to particular maker
	// of the algorithm. Contains only algorithms which don't require
	// any parameters.
	PolicyCombiningAlgs = map[string]PolicyCombiningAlgMaker{
		"firstapplicableeffect": makeFirstApplicableEffectPCA,
		"denyoverrides":         makeDenyOverridesPCA}

	// PolicyCombiningParamAlgs defines map of algorithm id to particular maker
	// of the algorithm. Contains only algorithms which require parameters.
	PolicyCombiningParamAlgs = map[string]PolicyCombiningAlgMaker{
		"mapper": makeMapperPCA}
)

// PolicySet represens PDP policy set (the set groups other policy sets and policies).
type PolicySet struct {
	id          string
	hidden      bool
	target      Target
	policies    []Evaluable
	obligations []AttributeAssignmentExpression
	algorithm   PolicyCombiningAlg
}

// NewPolicySet creates new instance of policy set with given id (or hidden),
// target, set of policy sets or policies, algorithm and obligations. To make
// instance of algorithm it uses one of makers from PolicyCombiningAlgs or
// PolicyCombiningParamAlgs and its parameters if it requires any.
func NewPolicySet(ID string, hidden bool, target Target, policies []Evaluable, makePCA PolicyCombiningAlgMaker, params interface{}, obligations []AttributeAssignmentExpression) *PolicySet {
	return &PolicySet{
		id:          ID,
		hidden:      hidden,
		target:      target,
		policies:    policies,
		obligations: obligations,
		algorithm:   makePCA(policies, params)}
}

func (p *PolicySet) describe() string {
	if pid, ok := p.GetID(); ok {
		return fmt.Sprintf("policy set %q", pid)
	}

	return "hidden policy set"
}

// GetID implements Evaluable interface and returns policy set id if policy set
// isn't hidden.
func (p *PolicySet) GetID() (string, bool) {
	return p.id, !p.hidden
}

// Calculate implements Evaluable interface and evaluates policy set for given
// request context.
func (p *PolicySet) Calculate(ctx *Context) Response {
	match, err := p.target.calculate(ctx)
	if err != nil {
		r := combineEffectAndStatus(err, p.algorithm.execute(p.policies, ctx))
		if r.status != nil {
			r.status = bindError(err, p.describe())
		}
		return r
	}

	if !match {
		return Response{EffectNotApplicable, nil, nil}
	}

	r := p.algorithm.execute(p.policies, ctx)
	if r.Effect == EffectDeny || r.Effect == EffectPermit {
		r.obligations = append(r.obligations, p.obligations...)
	}

	if r.status != nil {
		r.status = bindError(r.status, p.describe())
	}

	return r
}

// Append implements Evaluable interface and puts new policy set, policy or rule
// to the policy set or one of its children. Argument path should be empty
// to put policy set or policy to current policy set or contain ids of nested
// policy sets or policies to recurcively get to point where value of v argument
// can be appended. Value of v should be policy set or policy if path leads to
// policy set or rule if path leads to policy. Append can't put hidden item or
// any item to hidden policy set or policy.
func (p *PolicySet) Append(path []string, v interface{}) (Evaluable, error) {
	if p.hidden {
		return p, newHiddenPolicySetModificationError()
	}

	if len(path) > 0 {
		ID := path[0]

		child, err := p.getChild(ID)
		if err != nil {
			return p, bindError(err, p.id)
		}

		child, err = child.Append(path[1:], v)
		if err != nil {
			return p, bindError(err, p.id)
		}

		return p.putChild(child), nil
	}

	child, ok := v.(Evaluable)
	if !ok {
		return p, bindError(newInvalidPolicySetItemTypeError(v), p.id)
	}

	_, ok = child.GetID()
	if !ok {
		return p, bindError(newHiddenPolicyAppendError(), p.id)
	}

	return p.putChild(child), nil
}

// Delete implements Evaluable interface and removes item from the policy set
// or one of its children. Argument path should contain at least one string and
// should lead to item to delete. Delete can't remove an item from hidden policy
// set or policy.
func (p *PolicySet) Delete(path []string) (Evaluable, error) {
	if p.hidden {
		return p, newHiddenPolicySetModificationError()
	}

	if len(path) <= 0 {
		return p, bindError(newTooShortPathPolicySetModificationError(), p.id)
	}

	ID := path[0]

	if len(path) > 1 {
		child, err := p.getChild(ID)
		if err != nil {
			return p, bindError(err, p.id)
		}

		child, err = child.Delete(path[1:])
		if err != nil {
			return p, bindError(err, p.id)
		}

		return p.putChild(child), nil
	}

	r, err := p.delChild(ID)
	if err != nil {
		return p, bindError(err, p.id)
	}

	return r, nil
}

func (p *PolicySet) getChild(ID string) (Evaluable, error) {
	for _, child := range p.policies {
		if pID, ok := child.GetID(); ok && pID == ID {
			return child, nil
		}
	}

	return nil, newMissingPolicySetChildError(ID)
}

func (p *PolicySet) putChild(child Evaluable) Evaluable {
	ID, _ := child.GetID()

	for i, old := range p.policies {
		if pID, ok := old.GetID(); ok && pID == ID {
			policies := []Evaluable{}
			if i > 0 {
				policies = append(policies, p.policies[:i]...)
			}

			policies = append(policies, child)

			if i+1 < len(p.policies) {
				policies = append(policies, p.policies[i+1:]...)
			}

			algorithm := p.algorithm
			if m, ok := algorithm.(mapperPCA); ok {
				algorithm = m.add(ID, child, old)
			}

			return &PolicySet{
				id:          p.id,
				target:      p.target,
				policies:    policies,
				obligations: p.obligations,
				algorithm:   algorithm}
		}
	}

	policies := p.policies
	if policies == nil {
		policies = []Evaluable{child}
	} else {
		policies = append(policies, child)
	}

	algorithm := p.algorithm
	if m, ok := algorithm.(mapperPCA); ok {
		algorithm = m.add(ID, child, nil)
	}

	return &PolicySet{
		id:          p.id,
		target:      p.target,
		policies:    policies,
		obligations: p.obligations,
		algorithm:   algorithm}
}

func (p *PolicySet) delChild(ID string) (Evaluable, error) {
	for i, old := range p.policies {
		if pID, ok := old.GetID(); ok && pID == ID {
			policies := []Evaluable{}
			if i > 0 {
				policies = append(policies, p.policies[:i]...)
			}

			if i+1 < len(p.policies) {
				policies = append(policies, p.policies[i+1:]...)
			}

			algorithm := p.algorithm
			if m, ok := algorithm.(mapperPCA); ok {
				algorithm = m.del(ID, old)
			}

			return &PolicySet{
				id:          p.id,
				target:      p.target,
				policies:    policies,
				obligations: p.obligations,
				algorithm:   algorithm}, nil
		}
	}

	return nil, newMissingPolicySetChildError(ID)
}

type firstApplicableEffectPCA struct {
}

func makeFirstApplicableEffectPCA(policies []Evaluable, params interface{}) PolicyCombiningAlg {
	return firstApplicableEffectPCAInstance
}

func (a firstApplicableEffectPCA) execute(policies []Evaluable, ctx *Context) Response {
	for _, p := range policies {
		r := p.Calculate(ctx)
		if r.Effect != EffectNotApplicable {
			return r
		}
	}

	return Response{EffectNotApplicable, nil, nil}
}

type denyOverridesPCA struct {
}

func makeDenyOverridesPCA(policies []Evaluable, params interface{}) PolicyCombiningAlg {
	return denyOverridesPCAInstance
}

func (a denyOverridesPCA) describe() string {
	return "deny overrides"
}

func (a denyOverridesPCA) execute(policies []Evaluable, ctx *Context) Response {
	errs := []error{}
	obligations := make([]AttributeAssignmentExpression, 0)

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
