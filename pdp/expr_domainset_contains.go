package pdp

type functionSetOfDomainsContains struct {
	set   expression
	value expression
}

func (f functionSetOfDomainsContains) describe() string {
	return "contains"
}

func (f functionSetOfDomainsContains) calculate(ctx *Context) (attributeValue, error) {
	set, err := ctx.calculateSetOfDomainsExpression(f.set)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	value, err := ctx.calculateDomainExpression(f.value)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	_, ok := set.Get(value)
	return makeBooleanValue(ok), nil
}
