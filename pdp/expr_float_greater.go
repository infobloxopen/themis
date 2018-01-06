package pdp

import "fmt"

type functionFloatGreater struct {
	first  Expression
	second Expression
}

func makeFunctionFloatGreater(first, second Expression) Expression {
	return functionFloatGreater{
		first:  first,
		second: second,
	}
}

func makeFunctionFloatGreaterAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"greater\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatGreater(args[0], args[1])
}

func (f functionFloatGreater) GetResultType() int {
	return TypeBoolean
}

func (f functionFloatGreater) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), "equal")
	}

	second, err := ctx.calculateFloatExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), "equal")
	}

	return MakeBooleanValue(first > second), nil
}

func functionFloatGreaterValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeFloat || args[1].GetResultType() != TypeFloat {
		return nil
	}

	return makeFunctionFloatGreaterAlt
}
