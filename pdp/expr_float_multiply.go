package pdp

import "fmt"

type functionFloatMultiply struct {
	first  Expression
	second Expression
}

func makeFunctionFloatMultiply(first, second Expression) Expression {
	return functionFloatMultiply{
		first:  first,
		second: second,
	}
}

func makeFunctionFloatMultiplyAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"multiply\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatMultiply(args[0], args[1])
}

func (f functionFloatMultiply) GetResultType() int {
	return TypeFloat
}

func (f functionFloatMultiply) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatOrIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), "equal")
	}

	second, err := ctx.calculateFloatOrIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), "equal")
	}

	return MakeFloatValue(first * second), nil
}

func functionFloatMultiplyValidator(args []Expression) functionMaker {
	if len(args) != 2 ||
		(args[0].GetResultType() != TypeFloat && args[0].GetResultType() != TypeInteger) ||
		(args[1].GetResultType() != TypeFloat && args[1].GetResultType() != TypeInteger) {
		return nil
	}
	return makeFunctionFloatMultiplyAlt
}
