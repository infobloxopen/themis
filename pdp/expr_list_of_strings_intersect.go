package pdp

import (
	"fmt"
	"math"
	"sort"
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
	sort.Strings(first)
	sort.Strings(second)
	iLen, jLen := len(first), len(second)
	var res = make([]string, int(math.Max(float64(iLen), float64(jLen))))

	k := 0
	for i, j := 0, 0; i < iLen && j < jLen; {
		if first[i] > second[j] {
			if j < jLen {
				j++
			} else {
				break
			}
		} else if first[i] < second[j] {
			if i < iLen {
				i++
			} else {
				break
			}
		} else {
			val := first[i]
			res[k] = val
			k++
			if i < iLen {
				i++
			}
			if j < jLen {
				j++
			}
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
