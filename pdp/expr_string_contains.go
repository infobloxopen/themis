package pdp

import "strings"

type FunctionStringContainsType struct {
	Template ExpressionType
	Value    ExpressionType
}

func (f FunctionStringContainsType) getResultType() int {
	return DataTypeBoolean
}

func (f FunctionStringContainsType) calculate(ctx *Context) (AttributeValueType, error) {
	v := AttributeValueType{}

	template, err := f.Template.calculate(ctx)
	if err != nil {
		return v, err
	}

	tStr, err := ExtractStringValue(template, "template argument")
	if err != nil {
		return v, err
	}

	value, err := f.Value.calculate(ctx)
	if err != nil {
		return v, err
	}

	vStr, err := ExtractStringValue(value, "value argument")
	if err != nil {
		return v, err
	}

	v.DataType = DataTypeBoolean
	v.Value = strings.Contains(vStr, tStr)

	return v, nil
}

func makeFunctionStringContains(first ExpressionType, second ExpressionType) ExpressionType {
	return FunctionStringContainsType{first, second}
}

func makeFunctionStringContainsComm(args []ExpressionType) ExpressionType {
	return makeFunctionStringContains(args[0], args[1])
}

func checkerFunctionStringContains(args []ExpressionType) anyArgumentsFunctionType {
	if len(args) == 2 && args[0].getResultType() == DataTypeString && args[1].getResultType() == DataTypeString {
		return makeFunctionStringContainsComm
	}

	return nil
}
