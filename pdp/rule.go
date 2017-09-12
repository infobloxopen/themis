package pdp

import "fmt"

// Rule represents PDP rule (child or PDP policy).
type Rule struct {
	id          string
	hidden      bool
	target      Target
	condition   Expression
	effect      int
	obligations []AttributeAssignmentExpression
}

func makeConditionStatus(err boundError, effect int) Response {
	if effect == EffectDeny {
		return Response{EffectIndeterminateD, err, nil}
	}

	return Response{EffectIndeterminateP, err, nil}
}

// NewRule creates new instance of rule with given id (or hidden), target,
// condition, effect and obligations.
func NewRule(ID string, hidden bool, target Target, condition Expression, effect int, obligations []AttributeAssignmentExpression) *Rule {
	return &Rule{
		id:          ID,
		hidden:      hidden,
		target:      target,
		condition:   condition,
		effect:      effect,
		obligations: obligations}
}

func (r Rule) describe() string {
	if !r.hidden {
		return fmt.Sprintf("rule %q", r.id)
	}

	return "hidden rule"
}

// GetID returns rule id if the rule isn't hidden.
func (r Rule) GetID() (string, bool) {
	return r.id, !r.hidden
}

func (r Rule) calculate(ctx *Context) Response {
	match, boundErr := r.target.calculate(ctx)
	if boundErr != nil {
		return makeMatchStatus(bindError(boundErr, r.describe()), r.effect)
	}

	if !match {
		return Response{EffectNotApplicable, nil, nil}
	}

	if r.condition == nil {
		return Response{r.effect, nil, r.obligations}
	}

	c, err := ctx.calculateBooleanExpression(r.condition)
	if err != nil {
		return makeConditionStatus(bindError(bindError(err, "condition"), r.describe()), r.effect)
	}

	if !c {
		return Response{EffectNotApplicable, nil, nil}
	}

	return Response{r.effect, nil, r.obligations}
}
