package pdp

import "fmt"

type functionFloatEqual struct {
	first  Expression
	second Expression
}

func makeFunctionFloatEqual(first, second Expression) Expression {
	return functionFloatEqual{
		first:  first,
		second: second,
	}
}

func makeFunctionFloatEqualAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"equal\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatEqual(args[0], args[1])
}

func (f functionFloatEqual) GetResultType() int {
	return TypeBoolean
}

func (f functionFloatEqual) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), "equal")
	}

	second, err := ctx.calculateFloatExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), "equal")
	}

	return MakeBooleanValue(first == second), nil
}

func functionFloatEqualValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeFloat || args[1].GetResultType() != TypeFloat {
		return nil
	}

	return makeFunctionFloatEqualAlt
}
