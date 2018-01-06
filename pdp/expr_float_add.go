package pdp

import "fmt"

type functionFloatAdd struct {
	first  Expression
	second Expression
}

func makeFunctionFloatAdd(first, second Expression) Expression {
	return functionFloatAdd{
		first:  first,
		second: second,
	}
}

func makeFunctionFloatAddAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"add\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatAdd(args[0], args[1])
}

func (f functionFloatAdd) GetResultType() int {
	return TypeFloat
}

func (f functionFloatAdd) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), "equal")
	}

	second, err := ctx.calculateFloatExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), "equal")
	}

	return MakeFloatValue(first + second), nil
}

func functionFloatAddValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeFloat || args[1].GetResultType() != TypeFloat {
		return nil
	}

	return makeFunctionFloatAddAlt
}
