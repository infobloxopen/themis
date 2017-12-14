package pdp

import (
	"fmt"
	"regexp"
)

type functionStringRegexMatch struct {
	pattern Expression
	str     Expression
}

func makeFunctionStringRegexMatch(pattern, str Expression) Expression {
	return functionStringRegexMatch{
		pattern: pattern,
		str:     str}
}

func makeFunctionStringRegexMatchAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"regex-match\" for String needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionStringRegexMatch(args[0], args[1])
}

func (f functionStringRegexMatch) GetResultType() int {
	return TypeBoolean
}

func (f functionStringRegexMatch) describe() string {
	return "regex-match"
}

func (f functionStringRegexMatch) Calculate(ctx *Context) (AttributeValue, error) {
	pattern, err := ctx.calculateStringExpression(f.pattern)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "pattern"), f.describe())
	}

	str, err := ctx.calculateStringExpression(f.str)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "string"), f.describe())
	}

	ok, err := regexp.MatchString(pattern, str)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}
	return MakeBooleanValue(ok), nil
}

func functionStringRegexMatchValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeString || args[1].GetResultType() != TypeString {
		return nil
	}

	return makeFunctionStringRegexMatchAlt
}
