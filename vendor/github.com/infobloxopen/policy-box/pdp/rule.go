package pdp

import "fmt"

type RuleType struct {
	ID          string
	Target      TargetType
	Condition   ExpressionType
	Effect      int
	Obligations []AttributeAssignmentExpressionType
}

func conditionCalculate(c ExpressionType, ctx *Context) (AttributeValueType, error) {
	v, err := c.calculate(ctx)
	if err != nil {
		return v, err
	}

	if v.DataType != DataTypeBoolean {
		return v, fmt.Errorf("Expected boolean value as condition result but got %s",
			DataTypeNames[v.DataType])
	}

	return v, err
}

func (r RuleType) calculate(ctx *Context) ResponseType {
	match, err := r.Target.calculate(ctx)
	if err != nil {
		s := fmt.Sprintf("Match (%s): %s", r.ID, err)
		if r.Effect == EffectDeny {
			return ResponseType{EffectIndeterminateD, s, nil}
		}

		return ResponseType{EffectIndeterminateP, s, nil}
	}

	if !match {
		return ResponseType{EffectNotApplicable, "Ok", nil}
	}

	if r.Condition == nil {
		return ResponseType{r.Effect, "Ok", r.Obligations}
	}

	value, err := r.Condition.calculate(ctx)
	if err != nil {
		s := fmt.Sprintf("Condition (%s): %s", r.ID, err)
		if r.Effect == EffectDeny {
			return ResponseType{EffectIndeterminateD, s, nil}
		}

		return ResponseType{EffectIndeterminateP, s, nil}
	}

	if !value.Value.(bool) {
		return ResponseType{EffectNotApplicable, "Ok", nil}
	}

	return ResponseType{r.Effect, "Ok", r.Obligations}
}
