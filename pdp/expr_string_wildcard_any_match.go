package pdp

import "fmt"

const functionStringWildcardAnyMatchName = "wildcard-any-match"

type functionStringWildcardAnyMatch struct {
	patterns Expression
	str      Expression
}

func makeFunctionStringWildcardAnyMatch(patterns, str Expression) Expression {
	return functionStringWildcardAnyMatch{
		patterns: patterns,
		str:      str}
}

func makeFunctionStringWildcardAnyMatchAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function %q for Strings needs exactly two arguments but got %d",
			functionStringWildcardAnyMatchName, len(args)))
	}

	return makeFunctionStringWildcardAnyMatch(args[0], args[1])
}

func (f functionStringWildcardAnyMatch) GetResultType() int {
	return TypeBoolean
}

func (f functionStringWildcardAnyMatch) describe() string {
	return functionStringWildcardAnyMatchName
}

func (f functionStringWildcardAnyMatch) calculate(ctx *Context) (AttributeValue, error) {
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
			ok, err := wildcardMatch(p.Key, str)
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
			ok, err := wildcardMatch(p, str)
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

type functionStringWildcardMatchAny struct {
	pattern Expression
	strs    Expression
}

func makeFunctionStringWildcardMatchAny(pattern, strs Expression) Expression {
	return functionStringWildcardMatchAny{
		pattern: pattern,
		strs:    strs}
}

func makeFunctionStringWildcardMatchAnyAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function %q for Strings needs exactly two arguments but got %d",
			functionStringWildcardAnyMatchName, len(args)))
	}

	return makeFunctionStringWildcardMatchAny(args[0], args[1])
}

func (f functionStringWildcardMatchAny) GetResultType() int {
	return TypeBoolean
}

func (f functionStringWildcardMatchAny) describe() string {
	return functionStringWildcardAnyMatchName
}

func (f functionStringWildcardMatchAny) calculate(ctx *Context) (AttributeValue, error) {
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
			ok, err := wildcardMatch(pattern, s.Key)
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
			ok, err := wildcardMatch(pattern, s)
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

func functionStringWildcardAnyMatchValidator(args []Expression) functionMaker {
	if len(args) != 2 {
		return nil
	}

	pType := args[0].GetResultType()
	sType := args[1].GetResultType()

	if (pType == TypeSetOfStrings || pType == TypeListOfStrings) && sType == TypeString {
		return makeFunctionStringWildcardAnyMatchAlt
	}

	if pType == TypeString && (sType == TypeSetOfStrings || sType == TypeListOfStrings) {
		return makeFunctionStringWildcardMatchAnyAlt
	}

	return nil
}
