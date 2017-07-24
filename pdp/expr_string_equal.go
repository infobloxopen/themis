package pdp

type functionStringEqual struct {
	first  expression
	second expression
}

func (f functionStringEqual) describe() string {
	return "equal"
}

func (f functionStringEqual) calculate(ctx *Context) (attributeValue, error) {
	first, err := ctx.calculateStringExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateStringExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	return makeBooleanValue(first == second), nil
}
