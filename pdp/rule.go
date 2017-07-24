package pdp

import "fmt"

type rule struct {
	id          string
	target      target
	condition   expression
	effect      int
	obligations []attributeAssignmentExpression
}

func makeConditionStatus(err boundError, effect int) Response {
	if effect == EffectDeny {
		return Response{EffectIndeterminateD, err, nil}
	}

	return Response{EffectIndeterminateP, err, nil}
}

func (r rule) describe() string {
	return fmt.Sprintf("rule %q", r.id)
}

func (r rule) calculate(ctx *Context) Response {
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
