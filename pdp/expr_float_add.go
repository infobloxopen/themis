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

func (f functionFloatAdd) describe() string {
	return "add"
}

func (f functionFloatAdd) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateFloatOrIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}

	second, err := ctx.calculateFloatOrIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	return MakeFloatValue(first + second), nil
}

func functionFloatAddValidator(args []Expression) functionMaker {
	if len(args) != 2 ||
		(args[0].GetResultType() != TypeFloat && args[0].GetResultType() != TypeInteger) ||
		(args[1].GetResultType() != TypeFloat && args[1].GetResultType() != TypeInteger) {
		return nil
	}

	return makeFunctionFloatAddAlt
}
