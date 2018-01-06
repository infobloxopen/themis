package pdp

import "fmt"

type functionFloatSubtract struct {
	first  Expression
	second Expression
}

func makeFunctionFloatSubtract(first, second Expression) Expression {
	return functionFloatSubtract{
		first:  first,
		second: second,
	}
}

func makeFunctionFloatSubtractAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"subtract\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatSubtract(args[0], args[1])
}

func (f functionFloatSubtract) GetResultType() int {
	return TypeFloat
}

func (f functionFloatSubtract) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), "equal")
	}

	second, err := ctx.calculateFloatExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), "equal")
	}

	return MakeFloatValue(first - second), nil
}

func functionFloatSubtractValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeFloat || args[1].GetResultType() != TypeFloat {
		return nil
	}

	return makeFunctionFloatSubtractAlt
}
