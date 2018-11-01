package pdp

import "fmt"

type functionListOfStringsContains struct {
	list  Expression
	value Expression
}

func makeFunctionListOfStringsContains(list, value Expression) Expression {
	return functionListOfStringsContains{
		list:  list,
		value: value}
}

func makeFunctionListOfStringsContainsAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"contains\" for List Of Strings needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionListOfStringsContains(args[0], args[1])
}

func (f functionListOfStringsContains) GetResultType() Type {
	return TypeBoolean
}

func (f functionListOfStringsContains) describe() string {
	return "contains"
}

// Calculate implements Expression interface and returns calculated value
func (f functionListOfStringsContains) Calculate(ctx *Context) (AttributeValue, error) {
	list, err := ctx.calculateListOfStringsExpression(f.list)
	if err != nil {
		return UndefinedValue, bindError(err, f.describe())
	}

	s, err := ctx.calculateStringExpression(f.value)
	if err != nil {
		return UndefinedValue, bindError(err, f.describe())
	}

	for _, val := range list {
		if s == val {
			return MakeBooleanValue(true), nil
		}
	}
	return MakeBooleanValue(false), nil
}

func functionListOfStringsContainsValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeListOfStrings || args[1].GetResultType() != TypeString {
		return nil
	}

	return makeFunctionListOfStringsContainsAlt
}
