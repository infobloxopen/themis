package pdp

type FunctionSetOfNetworksContainsAddressType struct {
	Set   ExpressionType
	Value ExpressionType
}

func (f FunctionSetOfNetworksContainsAddressType) getResultType() int {
	return DataTypeBoolean
}

func (f FunctionSetOfNetworksContainsAddressType) calculate(ctx *Context) (AttributeValueType, error) {
	v := AttributeValueType{}

	set, err := f.Set.calculate(ctx)
	if err != nil {
		return v, err
	}

	s, err := ExtractSetOfNetworksValue(set, "first argument")
	if err != nil {
		return v, err
	}

	value, err := f.Value.calculate(ctx)
	if err != nil {
		return v, err
	}

	a, err := ExtractAddressValue(value, "second argument")
	if err != nil {
		return v, err
	}

	v.DataType = DataTypeBoolean
	for _, n := range s {
		if n.Contains(a) {
			v.Value = true
			return v, nil
		}
	}

	v.Value = false
	return v, nil
}

func makeFunctionSetOfNetworksContainsAddress(first ExpressionType, second ExpressionType) ExpressionType {
	return FunctionSetOfNetworksContainsAddressType{first, second}
}

func makeFunctionSetOfNetworksContainsAddressComm(args []ExpressionType) ExpressionType {
	return makeFunctionSetOfNetworksContainsAddress(args[0], args[1])
}

func checkerFunctionSetOfNetworksContainsAddress(args []ExpressionType) anyArgumentsFunctionType {
	if len(args) == 2 && args[0].getResultType() == DataTypeSetOfNetworks && args[1].getResultType() == DataTypeAddress {
		return makeFunctionSetOfNetworksContainsAddressComm
	}

	return nil
}
