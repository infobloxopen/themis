package pdp

import "fmt"

type functionIntegerAdd struct {
	first  Expression
	second Expression
}

func makeFunctionIntegerAdd(first, second Expression) Expression {
	return functionIntegerAdd{
		first:  first,
		second: second,
	}
}

func makeFunctionIntegerAddAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"add\" for Integer needs exactly two arguments but got %d", len(args)))
	}

	return functionIntegerAdd{
		first:  args[0],
		second: args[1],
	}
}

func (f functionIntegerAdd) GetResultType() int {
	return TypeInteger
}

func (f functionIntegerAdd) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), "equal")
	}

	second, err := ctx.calculateIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), "equal")
	}

	return MakeIntegerValue(first+second), nil
}

func functionIntegerAddValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeInteger || args[1].GetResultType() != TypeInteger {
		return nil
	}

	return makeFunctionIntegerAddAlt
}
