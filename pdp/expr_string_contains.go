package pdp

import "strings"

type functionStringContains struct {
	str    expression
	substr expression
}

func (f functionStringContains) describe() string {
	return "contains"
}

func (f functionStringContains) calculate(ctx *Context) (attributeValue, error) {
	str, err := ctx.calculateStringExpression(f.str)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "string argument"), f.describe())
	}

	substr, err := ctx.calculateStringExpression(f.substr)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "substring argument"), f.describe())
	}

	return makeBooleanValue(strings.Contains(str, substr)), nil
}
