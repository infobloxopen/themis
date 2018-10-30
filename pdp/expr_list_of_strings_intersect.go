package pdp

import "fmt"

type functionListOfStringsIntersect struct {
	first Expression
	second Expression
}

func makeFunctionListOfStringsIntersect(first, second Expression) Expression {
	return functionListOfStringsIntersect{
		first: first,
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
	// FIXME: room for improvement here...
	var res []string
	for _, f := range first {
		for _, s := range second {
			if f == s {
				res = append(res, f)
			}
		}
	}
	return MakeListOfStringsValue(res), nil
}

func functionListOfStringsIntersectValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeListOfStrings || args[1].GetResultType() != TypeListOfStrings {
		return nil
	}
	return makeFunctionListOfStringsIntersectAlt
}