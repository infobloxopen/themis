package pdp

import (
	"fmt"
)

type functionSetOfStringsLen struct {
	e Expression
}

func makeFunctionSetOfStringsLen(e Expression) Expression {
	return functionSetOfStringsLen{e: e}
}

func makeFunctionSetOfStringsLenAlt(args []Expression) Expression {
	if len(args) != 1 {
		panic(fmt.Errorf("function \"len\" for Set of Strings needs exactly one arguments but got %d", len(args)))
	}
	return makeFunctionSetOfStringsLen(args[0])
}

func (f functionSetOfStringsLen) GetResultType() Type {
	return TypeInteger
}

func (f functionSetOfStringsLen) describe() string {
	return "len"
}

// Calculate implements Expression interface and returns calculated value
func (f functionSetOfStringsLen) Calculate(ctx *Context) (AttributeValue, error) {
	set, err := ctx.calculateSetOfStringsExpression(f.e)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "argument"), f.describe())
	}

	l := 0
	for range set.Enumerate() {
		l++
	}

	return MakeIntegerValue(int64(l)), nil
}

func functionSetOfStringsLenValidator(args []Expression) functionMaker {
	if len(args) != 1 || args[0].GetResultType() != TypeSetOfStrings {
		return nil
	}
	return makeFunctionSetOfStringsLenAlt
}
