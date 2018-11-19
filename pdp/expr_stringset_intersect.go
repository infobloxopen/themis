package pdp

import (
	"fmt"

	"github.com/infobloxopen/go-trees/strtree"
)

type functionSetOfStringsIntersect struct {
	first  Expression
	second Expression
}

func makeFunctionSetOfStringsIntersect(first, second Expression) Expression {
	return functionSetOfStringsIntersect{
		first:  first,
		second: second}
}

func makeFunctionSetOfStringsIntersectAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"intersect\" for Set of Strings needs exactly two arguments but got %d", len(args)))
	}
	return makeFunctionSetOfStringsIntersect(args[0], args[1])
}

func (f functionSetOfStringsIntersect) GetResultType() Type {
	return TypeSetOfStrings
}

func (f functionSetOfStringsIntersect) describe() string {
	return "intersect"
}

// Calculate implements Expression interface and returns calculated value
func (f functionSetOfStringsIntersect) Calculate(ctx *Context) (AttributeValue, error) {
	firstSet, err := ctx.calculateSetOfStringsExpression(f.first)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}
	secondSet, err := ctx.calculateSetOfStringsExpression(f.second)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	res := strtree.NewTree()
	first := firstSet.Enumerate()
	second := secondSet.Enumerate()
	prev := ""

	for f, fok, s, sok := func() (
		f strtree.Pair, fok bool,
		s strtree.Pair, sok bool) {
		f, fok = <-first
		s, sok = <-second
		return
	}(); fok && sok; {
		fkey, skey := f.Key, s.Key
		if fkey > skey {
			s, sok = <-second
		} else if fkey < skey {
			f, fok = <-first
		} else {
			if fkey != prev || res.IsEmpty() {
				res.InplaceInsert(fkey, 0)
				prev = fkey
			}
			f, fok = <-first
			s, sok = <-second
		}
	}

	return MakeSetOfStringsValue(res), nil
}

func functionSetOfStringsIntersectValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeSetOfStrings || args[1].GetResultType() != TypeSetOfStrings {
		return nil
	}
	return makeFunctionSetOfStringsIntersectAlt
}
