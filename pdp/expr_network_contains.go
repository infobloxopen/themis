package pdp

type FunctionNetworkContainsAddressType struct {
	Network ExpressionType
	Address ExpressionType
}
func (f FunctionNetworkContainsAddressType) getResultType() int {
	return DataTypeBoolean
}

func (f FunctionNetworkContainsAddressType) calculate(ctx *Context) (AttributeValueType, error) {
	v := AttributeValueType{}

	network, err := f.Network.calculate(ctx)
	if err != nil {
		return v, err
	}

	n, err := ExtractNetworkValue(network, "first argument")
	if err != nil {
		return v, err
	}

	address, err := f.Address.calculate(ctx)
	if err != nil {
		return v, err
	}

	a, err := ExtractAddressValue(address, "second argument")
	if err != nil {
		return v, err
	}

	v.DataType = DataTypeBoolean
	v.Value = n.Contains(a)

	return v, nil
}
func makeFunctionNetworkContainsAddress(first ExpressionType, second ExpressionType) ExpressionType {
	return FunctionNetworkContainsAddressType{first, second}
}

func makeFunctionNetworkContainsAddressComm(args []ExpressionType) ExpressionType {
	return makeFunctionNetworkContainsAddress(args[0], args[1])
}

func checkerFunctionNetworkContainsAddress(args []ExpressionType) anyArgumentsFunctionType {
	if len(args) == 2 && args[0].getResultType() == DataTypeNetwork && args[1].getResultType() == DataTypeAddress {
		return makeFunctionNetworkContainsAddressComm
	}

	return nil
}
