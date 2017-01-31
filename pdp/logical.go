package pdp

import "fmt"

type FunctionBooleanNotType struct {
	Arg ExpressionType
}

type FunctionBooleanOrType struct {
	Args []ExpressionType
}

type FunctionBooleanAndType struct {
	Args []ExpressionType
}

func (f FunctionBooleanNotType) getResultType() int {
	return DataTypeBoolean
}

func (f FunctionBooleanNotType) calculate(ctx *Context) (AttributeValueType, error) {
	v := AttributeValueType{}

	r, err := f.Arg.calculate(ctx)
	if err != nil {
		return v, err
	}

	a, err := ExtractBooleanValue(r, "argument")
	if err != nil {
		return v, err
	}

	v.DataType = DataTypeBoolean
	v.Value = !a

	return v, nil
}

func (f FunctionBooleanOrType) getResultType() int {
	return DataTypeBoolean
}

func (f FunctionBooleanOrType) calculate(ctx *Context) (AttributeValueType, error) {
	v := AttributeValueType{}

	for i, arg := range f.Args {
		r, err := arg.calculate(ctx)
		if err != nil {
			return v, err
		}

		a, err := ExtractBooleanValue(r, fmt.Sprintf("argument %d", i))
		if err != nil {
			return v, err
		}

		if a {
			v.DataType = DataTypeBoolean
			v.Value = true
			return v, nil
		}
	}

	v.DataType = DataTypeBoolean
	v.Value = false

	return v, nil
}

func (f FunctionBooleanAndType) getResultType() int {
	return DataTypeBoolean
}

func (f FunctionBooleanAndType) calculate(ctx *Context) (AttributeValueType, error) {
	v := AttributeValueType{}

	for i, arg := range f.Args {
		r, err := arg.calculate(ctx)
		if err != nil {
			return v, err
		}

		a, err := ExtractBooleanValue(r, fmt.Sprintf("argument %d", i))
		if err != nil {
			return v, err
		}

		if !a {
			v.DataType = DataTypeBoolean
			v.Value = false
			return v, nil
		}
	}

	v.DataType = DataTypeBoolean
	v.Value = true

	return v, nil
}

func makeFunctionBooleanNotComm(args []ExpressionType) ExpressionType {
	return FunctionBooleanNotType{args[0]}
}

func checkerFunctionBooleanNot(args []ExpressionType) anyArgumentsFunctionType {
	if len(args) == 1 && args[0].getResultType() == DataTypeBoolean {
		return makeFunctionBooleanNotComm
	}

	return nil
}

func makeFunctionBooleanOrComm(args []ExpressionType) ExpressionType {
	return FunctionBooleanOrType{args}
}

func checkerFunctionBooleanOr(args []ExpressionType) anyArgumentsFunctionType {
	if len(args) < 2 {
		return nil
	}

	for _, a := range args {
		if a.getResultType() != DataTypeBoolean {
			return nil
		}
	}

	return makeFunctionBooleanOrComm
}

func makeFunctionBooleanAndComm(args []ExpressionType) ExpressionType {
	return FunctionBooleanAndType{args}
}

func checkerFunctionBooleanAnd(args []ExpressionType) anyArgumentsFunctionType {
	if len(args) < 2 {
		return nil
	}

	for _, a := range args {
		if a.getResultType() != DataTypeBoolean {
			return nil
		}
	}

	return makeFunctionBooleanAndComm
}
