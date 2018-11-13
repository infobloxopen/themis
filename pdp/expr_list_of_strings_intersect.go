package pdp

import (
	"fmt"
	"math"
)

type functionListOfStringsIntersect struct {
	first  Expression
	second Expression
}

func makeFunctionListOfStringsIntersect(first, second Expression) Expression {
	return functionListOfStringsIntersect{
		first:  first,
		second: second}
}

func makeFunctionListOfStringsIntersectAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"intersect\" for List of Strings needs exactly two arguments but got %d", len(args)))
	}
	return makeFunctionListOfStringsIntersect(args[0], args[1])
}

func (f functionListOfStringsIntersect) GetResultType() Type {
	return TypeListOfStrings
}

func (f functionListOfStringsIntersect) describe() string {
	return "intersect"
}

// Calculate implements Expression interface and returns calculated value
func (f functionListOfStringsIntersect) Calculate(ctx *Context) (AttributeValue, error) {
	first, err := ctx.calculateListOfStringsExpression(f.first)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}
	second, err := ctx.calculateListOfStringsExpression(f.second)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	values := make(map[string]bool)
	for _, f := range first {
		values[f] = false
	}

	res := make([]string, int(math.Min(float64(len(first)), float64(len(second)))))

	k := 0
	for _, s := range second {
		if found, ok := values[s]; ok && !found {
			values[s] = true
			res[k] = s
			k++
		}
	}

	return MakeListOfStringsValue(res[:k]), nil
}

func functionListOfStringsIntersectValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeListOfStrings || args[1].GetResultType() != TypeListOfStrings {
		return nil
	}
	return makeFunctionListOfStringsIntersectAlt
}
