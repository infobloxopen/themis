package pdp

import "fmt"

type functionIntegerGreater struct {
	first  Expression
	second Expression
}

func makeFunctionIntegerGreater(first, second Expression) Expression {
	return functionIntegerGreater{
		first:  first,
		second: second,
	}
}

func makeFunctionIntegerGreaterAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"greater\" for Integer needs exactly two arguments but got %d", len(args)))
	}

	return functionIntegerGreater{
		first:  args[0],
		second: args[1],
	}
}

func (f functionIntegerGreater) GetResultType() int {
	return TypeBoolean
}

func (f functionIntegerGreater) calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateIntegerExpression(f.first)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "first argument"), "equal")
	}

	second, err := ctx.calculateIntegerExpression(f.second)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "second argument"), "equal")
	}

	return MakeBooleanValue(first > second), nil
}

func functionIntegerGreaterValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeInteger || args[1].GetResultType() != TypeInteger {
		return nil
	}

	return makeFunctionIntegerGreaterAlt
}
