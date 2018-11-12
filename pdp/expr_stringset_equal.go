package pdp

import (
	"fmt"

	"github.com/infobloxopen/go-trees/strtree"
)

type functionSetOfStringsEqual struct {
	first  Expression
	second Expression
}

func makeFunctionSetOfStringsEqual(first, second Expression) Expression {
	return functionSetOfStringsEqual{
		first:  first,
		second: second}
}

func makeFunctionSetOfStringsEqualAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"equal\" for Set of Strings needs exactly two arguments but got %d", len(args)))
	}
	return makeFunctionSetOfStringsEqual(args[0], args[1])
}

func (f functionSetOfStringsEqual) GetResultType() Type {
	return TypeBoolean
}

func (f functionSetOfStringsEqual) describe() string {
	return "equal"
}

// Calculate implements Expression interface and returns calculated value
func (f functionSetOfStringsEqual) Calculate(ctx *Context) (AttributeValue, error) {
	firstSet, err := ctx.calculateSetOfStringsExpression(f.first)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "first argument"), f.describe())
	}
	secondSet, err := ctx.calculateSetOfStringsExpression(f.second)
	if err != nil {
		return UndefinedValue, bindError(bindError(err, "second argument"), f.describe())
	}

	first := firstSet.Enumerate()
	second := secondSet.Enumerate()
	fok, sok := true, true

	for f, s := *new(strtree.Pair), *new(strtree.Pair); fok && sok; {
		if f.Key != s.Key {
			return MakeBooleanValue(false), nil
		}
		f, fok = <-first
		s, sok = <-second
	}

	return MakeBooleanValue(fok == sok), nil
}

func functionSetOfStringsEqualValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeSetOfStrings || args[1].GetResultType() != TypeSetOfStrings {
		return nil
	}
	return makeFunctionSetOfStringsEqualAlt
}
