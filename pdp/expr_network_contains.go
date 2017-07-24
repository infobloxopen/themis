package pdp

type functionNetworkContainsAddress struct {
	network expression
	address expression
}

func (f functionNetworkContainsAddress) describe() string {
	return "contains"
}

func (f functionNetworkContainsAddress) calculate(ctx *Context) (attributeValue, error) {
	n, err := ctx.calculateNetworkExpression(f.network)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	a, err := ctx.calculateAddressExpression(f.address)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}

	return makeBooleanValue(n.Contains(a)), nil
}
