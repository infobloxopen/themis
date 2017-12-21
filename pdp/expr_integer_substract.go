package pdp

import "fmt"

type functionIntegerSubtract struct {
	first  Expression
	second Expression
}

func makeFunctionIntegerSubtract(first, second Expression) Expression {
	return functionIntegerSubtract{
		first:  first,
		second: second,
	}
}

func makeFunctionIntegerSubtractAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"subtract\" for Integer needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionIntegerSubtract(args[0], args[1])
}

func (f functionIntegerSubtract) GetResultType() int {
	return TypeInteger
}

func (f functionIntegerSubtract) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), "equal")
	}

	second, err := ctx.calculateIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), "equal")
	}

	return MakeIntegerValue(first - second), nil
}

func functionIntegerSubtractValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeInteger || args[1].GetResultType() != TypeInteger {
		return nil
	}

	return makeFunctionIntegerSubtractAlt
}
