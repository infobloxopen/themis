package pdp

import "fmt"

type FunctionSetOfStringContainsType struct {
	Set   ExpressionType
	Value ExpressionType
}

func (f FunctionSetOfStringContainsType) describe() string {
	return fmt.Sprintf("\"%s\".%s(%s)", yastTagDataTypeSetOfStrings, yastExpressionContains, yastTagDataTypeString)
}

func (f FunctionSetOfStringContainsType) getResultType() int {
	return DataTypeBoolean
}

func (f FunctionSetOfStringContainsType) calculate(ctx *Context) (AttributeValueType, error) {
	v := AttributeValueType{}

	set, err := f.Set.calculate(ctx)
	if err != nil {
		return v, err
	}

	s, err := ExtractSetOfStringsValue(set, "first argument")
	if err != nil {
		return v, err
	}

	value, err := f.Value.calculate(ctx)
	if err != nil {
		return v, err
	}

	vStr, err := ExtractStringValue(value, "second argument")
	if err != nil {
		return v, err
	}

	v.DataType = DataTypeBoolean
	v.Value = s[vStr]

	return v, nil
}

func makeFunctionSetOfStringContains(first ExpressionType, second ExpressionType) ExpressionType {
	return FunctionSetOfStringContainsType{first, second}
}

func makeFunctionSetOfStringContainsComm(args []ExpressionType) ExpressionType {
	return makeFunctionSetOfStringContains(args[0], args[1])
}

func checkerFunctionSetOfStringContains(args []ExpressionType) anyArgumentsFunctionType {
	if len(args) == 2 && args[0].getResultType() == DataTypeSetOfStrings && args[1].getResultType() == DataTypeString {
		return makeFunctionSetOfStringContainsComm
	}

	return nil
}
