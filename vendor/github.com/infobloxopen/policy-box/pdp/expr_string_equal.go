package pdp

import "fmt"

type FunctionStringEqualType struct {
	First  ExpressionType
	Second ExpressionType
}

func (f FunctionStringEqualType) describe() string {
	return fmt.Sprintf("%s(%s, %s)", yastExpressionEqual, yastTagDataTypeString, yastTagDataTypeString)
}

func (f FunctionStringEqualType) getResultType() int {
	return DataTypeBoolean
}

func (f FunctionStringEqualType) calculate(ctx *Context) (AttributeValueType, error) {
	v := AttributeValueType{}

	first, err := f.First.calculate(ctx)
	if err != nil {
		return v, err
	}

	firstStr, err := ExtractStringValue(first, "first argument")
	if err != nil {
		return v, err
	}

	second, err := f.Second.calculate(ctx)
	if err != nil {
		return v, err
	}

	secondStr, err := ExtractStringValue(second, "second argument")
	if err != nil {
		return v, err
	}

	v.DataType = DataTypeBoolean
	v.Value = firstStr == secondStr

	return v, nil
}

func makeFunctionStringEqual(first ExpressionType, second ExpressionType) ExpressionType {
	return FunctionStringEqualType{first, second}
}

func makeFunctionStringEqualComm(args []ExpressionType) ExpressionType {
	return makeFunctionStringEqual(args[0], args[1])
}

func checkerFunctionStringEqual(args []ExpressionType) anyArgumentsFunctionType {
	if len(args) == 2 && args[0].getResultType() == DataTypeString && args[1].getResultType() == DataTypeString {
		return makeFunctionStringEqualComm
	}

	return nil
}
