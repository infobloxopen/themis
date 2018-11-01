package pdp

import "fmt"

type functionListOfStringsLen struct {
	e Expression
}

func makeFunctionListOfStringsLen(e Expression) Expression {
	return functionListOfStringsLen{e: e}
}

func makeFunctionListOfStringsLenAlt(args []Expression) Expression {
	if len(args) != 1 {
		panic(fmt.Errorf("function \"len\" for List of Strings needs exactly one arguments but got %d", len(args)))
	}
	return makeFunctionListOfStringsLen(args[0])
}

func (f functionListOfStringsLen) GetResultType() Type {
	return TypeInteger
}

func (f functionListOfStringsLen) describe() string {
	return "len"
}

// Calculate implements Expression interface and returns calculated value
func (f functionListOfStringsLen) Calculate(ctx *Context) (AttributeValue, error) {
	s, err := ctx.calculateListOfStringsExpression(f.e)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "argument"), f.describe())
	}
	return MakeIntegerValue(int64(len(s))), nil
}

func functionListOfStringsLenValidator(args []Expression) functionMaker {
	if len(args) != 1 || args[0].GetResultType() != TypeListOfStrings {
		return nil
	}
	return makeFunctionListOfStringsLenAlt
}
