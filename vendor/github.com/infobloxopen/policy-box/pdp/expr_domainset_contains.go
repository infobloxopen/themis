package pdp

import "fmt"

type FunctionSetOfDomainsContainsType struct {
	Set   ExpressionType
	Value ExpressionType
}

func (f FunctionSetOfDomainsContainsType) describe() string {
	return fmt.Sprintf("\"%s\".%s(%s)", yastTagDataTypeSetOfDomains, yastExpressionContains, yastTagDataTypeDomain)
}

func (f FunctionSetOfDomainsContainsType) getResultType() int {
	return DataTypeBoolean
}

func (f FunctionSetOfDomainsContainsType) calculate(ctx *Context) (AttributeValueType, error) {
	v := AttributeValueType{}

	set, err := f.Set.calculate(ctx)
	if err != nil {
		return v, err
	}

	s, err := ExtractSetOfDomainsValue(set, "first argument")
	if err != nil {
		return v, err
	}

	value, err := f.Value.calculate(ctx)
	if err != nil {
		return v, err
	}

	d, err := ExtractDomainValue(value, "second argument")
	if err != nil {
		return v, err
	}

	v.DataType = DataTypeBoolean
	v.Value = s.Contains(d)

	return v, nil
}

func makeFunctionSetOfDomainsContains(first ExpressionType, second ExpressionType) ExpressionType {
	return FunctionSetOfDomainsContainsType{first, second}
}

func makeFunctionSetOfDomainsContainsComm(args []ExpressionType) ExpressionType {
	return makeFunctionSetOfDomainsContains(args[0], args[1])
}

func checkerFunctionSetOfDomainsContains(args []ExpressionType) anyArgumentsFunctionType {
	if len(args) == 2 && args[0].getResultType() == DataTypeSetOfDomains && args[1].getResultType() == DataTypeDomain {
		return makeFunctionSetOfDomainsContainsComm
	}

	return nil
}
