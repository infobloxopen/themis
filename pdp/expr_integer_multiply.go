package pdp

import "fmt"

type functionIntegerMultiply struct {
	first  Expression
	second Expression
}

func makeFunctionIntegerMultiply(first, second Expression) Expression {
	return functionIntegerMultiply{
		first:  first,
		second: second,
	}
}

func makeFunctionIntegerMultiplyAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"multiply\" for Integer needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionIntegerMultiply(args[0], args[1])
}

func (f functionIntegerMultiply) GetResultType() int {
	return TypeInteger
}

func (f functionIntegerMultiply) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), "equal")
	}

	second, err := ctx.calculateIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), "equal")
	}

	return MakeIntegerValue(first * second), nil
}

func functionIntegerMultiplyValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeInteger || args[1].GetResultType() != TypeInteger {
		return nil
	}

	return makeFunctionIntegerMultiplyAlt
}
