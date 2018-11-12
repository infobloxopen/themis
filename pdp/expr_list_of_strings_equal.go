package pdp

import (
	"fmt"
)

type functionListOfStringsEqual struct {
	first  Expression
	second Expression
}

func makeFunctionListOfStringsEqual(first, second Expression) Expression {
	return functionListOfStringsEqual{
		first:  first,
		second: second}
}

func makeFunctionListOfStringsEqualAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"equal\" for List of Strings needs exactly two arguments but got %d", len(args)))
	}
	return makeFunctionListOfStringsEqual(args[0], args[1])
}

func (f functionListOfStringsEqual) GetResultType() Type {
	return TypeBoolean
}

func (f functionListOfStringsEqual) describe() string {
	return "equal"
}

// Calculate implements Expression interface and returns calculated value
func (f functionListOfStringsEqual) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateListOfStringsExpression(f.first)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}
	second, err := ctx.calculateListOfStringsExpression(f.second)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	if len(first) != len(second) {
		return MakeBooleanValue(false), nil
	}

	for i := range first {
		if first[i] != second[i] {
			return MakeBooleanValue(false), nil
		}
	}

	return MakeBooleanValue(true), nil
}

func functionListOfStringsEqualValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeListOfStrings || args[1].GetResultType() != TypeListOfStrings {
		return nil
	}
	return makeFunctionListOfStringsEqualAlt
}
