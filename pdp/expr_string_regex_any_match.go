package pdp

import (
	"fmt"
	"regexp"
)

const functionStringRegexAnyMatchName = "regex-any-match"

type functionStringRegexAnyMatch struct {
	patterns Expression
	str      Expression
}

func makeFunctionStringRegexAnyMatch(patterns, str Expression) Expression {
	return functionStringRegexAnyMatch{
		patterns: patterns,
		str:      str}
}

func makeFunctionStringRegexAnyMatchAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function %q for Strings needs exactly two arguments but got %d",
			functionStringRegexAnyMatchName, len(args)))
	}

	return makeFunctionStringRegexAnyMatch(args[0], args[1])
}

func (f functionStringRegexAnyMatch) GetResultType() int {
	return TypeBoolean
}

func (f functionStringRegexAnyMatch) describe() string {
	return functionStringRegexAnyMatchName
}

func (f functionStringRegexAnyMatch) Calculate(ctx *Context) (AttributeValue, error) {
	pType := f.patterns.GetResultType()
	switch pType {
	case TypeSetOfStrings:
		patterns, err := ctx.calculateSetOfStringsExpression(f.patterns)
		if err != nil {
			return undefinedValue, bindError(bindError(err, "patterns"), f.describe())
		}

		str, err := ctx.calculateStringExpression(f.str)
		if err != nil {
			return undefinedValue, bindError(bindError(err, "string"), f.describe())
		}

		for p := range patterns.Enumerate() {
			ok, err := regexp.MatchString(p.Key, str)
			if err != nil {
				return undefinedValue, bindError(err, f.describe())
			}

			if ok {
				return MakeBooleanValue(true), nil
			}
		}

		return MakeBooleanValue(false), nil

	case TypeListOfStrings:
		patterns, err := ctx.calculateListOfStringsExpression(f.patterns)
		if err != nil {
			return undefinedValue, bindError(bindError(err, "patterns"), f.describe())
		}

		str, err := ctx.calculateStringExpression(f.str)
		if err != nil {
			return undefinedValue, bindError(bindError(err, "string"), f.describe())
		}

		for _, p := range patterns {
			ok, err := regexp.MatchString(p, str)
			if err != nil {
				return undefinedValue, bindError(err, f.describe())
			}

			if ok {
				return MakeBooleanValue(true), nil
			}
		}

		return MakeBooleanValue(false), nil

	}

	return undefinedValue,
		bindError(
			bindError(newInvalidContainerArgType(pType, TypeSetOfStrings, TypeListOfStrings),
				"patterns",
			),
			f.describe(),
		)
}

type functionStringRegexMatchAny struct {
	pattern Expression
	strs    Expression
}

func makeFunctionStringRegexMatchAny(pattern, strs Expression) Expression {
	return functionStringRegexMatchAny{
		pattern: pattern,
		strs:    strs}
}

func makeFunctionStringRegexMatchAnyAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function %q for Strings needs exactly two arguments but got %d",
			functionStringRegexAnyMatchName, len(args)))
	}

	return makeFunctionStringRegexMatchAny(args[0], args[1])
}

func (f functionStringRegexMatchAny) GetResultType() int {
	return TypeBoolean
}

func (f functionStringRegexMatchAny) describe() string {
	return functionStringRegexAnyMatchName
}

func (f functionStringRegexMatchAny) Calculate(ctx *Context) (AttributeValue, error) {
	pattern, err := ctx.calculateStringExpression(f.pattern)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "pattern"), f.describe())
	}

	sType := f.strs.GetResultType()
	switch sType {
	case TypeSetOfStrings:
		strs, err := ctx.calculateSetOfStringsExpression(f.strs)
		if err != nil {
			return undefinedValue, bindError(bindError(err, "strings"), f.describe())
		}

		for s := range strs.Enumerate() {
			ok, err := regexp.MatchString(pattern, s.Key)
			if err != nil {
				return undefinedValue, bindError(err, f.describe())
			}

			if ok {
				return MakeBooleanValue(true), nil
			}
		}

		return MakeBooleanValue(false), nil

	case TypeListOfStrings:
		strs, err := ctx.calculateListOfStringsExpression(f.strs)
		if err != nil {
			return undefinedValue, bindError(bindError(err, "strings"), f.describe())
		}

		for _, s := range strs {
			ok, err := regexp.MatchString(pattern, s)
			if err != nil {
				return undefinedValue, bindError(err, f.describe())
			}

			if ok {
				return MakeBooleanValue(true), nil
			}
		}

		return MakeBooleanValue(false), nil

	}

	return undefinedValue,
		bindError(
			bindError(newInvalidContainerArgType(sType, TypeSetOfStrings, TypeListOfStrings),
				"strings",
			),
			f.describe(),
		)
}

func functionStringRegexAnyMatchValidator(args []Expression) functionMaker {
	if len(args) != 2 {
		return nil
	}

	pType := args[0].GetResultType()
	sType := args[1].GetResultType()

	if (pType == TypeSetOfStrings || pType == TypeListOfStrings) && sType == TypeString {
		return makeFunctionStringRegexAnyMatchAlt
	}

	if pType == TypeString && (sType == TypeSetOfStrings || sType == TypeListOfStrings) {
		return makeFunctionStringRegexMatchAnyAlt
	}

	return nil
}
