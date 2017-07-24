package pdp

type functionSetOfNetworksContainsAddress struct {
	set   expression
	value expression
}

func (f functionSetOfNetworksContainsAddress) describe() string {
	return "contains"
}

func (f functionSetOfNetworksContainsAddress) calculate(ctx *Context) (attributeValue, error) {
	set, err := ctx.calculateSetOfNetworksExpression(f.set)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	a, err := ctx.calculateAddressExpression(f.value)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	_, ok := set.GetByIP(a)
	return makeBooleanValue(ok), nil
}
