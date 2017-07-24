package pdp

import "fmt"

type functionBooleanNot struct {
	arg expression
}

type functionBooleanOr struct {
	args []expression
}

type functionBooleanAnd struct {
	args []expression
}

func (f functionBooleanNot) describe() string {
	return "not"
}

func (f functionBooleanNot) calculate(ctx *Context) (attributeValue, error) {
	a, err := ctx.calculateBooleanExpression(f.arg)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	return makeBooleanValue(!a), nil
}

func (f functionBooleanOr) describe() string {
	return "or"
}

func (f functionBooleanOr) calculate(ctx *Context) (attributeValue, error) {
	for i, arg := range f.args {
		a, err := ctx.calculateBooleanExpression(arg)
		if err != nil {
			return undefinedValue, bindError(bindError(err, fmt.Sprintf("argument %d", i)), f.describe())
		}

		if a {
			return makeBooleanValue(true), nil
		}
	}

	return makeBooleanValue(false), nil
}

func (f functionBooleanAnd) describe() string {
	return "and"
}

func (f functionBooleanAnd) calculate(ctx *Context) (attributeValue, error) {
	for i, arg := range f.args {
		a, err := ctx.calculateBooleanExpression(arg)
		if err != nil {
			return undefinedValue, bindError(bindError(err, fmt.Sprintf("argument %d", i)), f.describe())
		}

		if !a {
			return makeBooleanValue(false), nil
		}
	}

	return makeBooleanValue(true), nil
}
