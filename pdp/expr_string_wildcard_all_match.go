package pdp

import "fmt"

const functionStringWildcardAllMatchName = "wildcard-all-match"

type functionStringWildcardAllMatch struct {
	patterns Expression
	str      Expression
}

func makeFunctionStringWildcardAllMatch(patterns, str Expression) Expression {
	return functionStringWildcardAllMatch{
		patterns: patterns,
		str:      str}
}

func makeFunctionStringWildcardAllMatchAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function %q for Strings needs exactly two arguments but got %d",
			functionStringWildcardAllMatchName, len(args)))
	}

	return makeFunctionStringWildcardAllMatch(args[0], args[1])
}

func (f functionStringWildcardAllMatch) GetResultType() int {
	return TypeBoolean
}

func (f functionStringWildcardAllMatch) describe() string {
	return functionStringWildcardAllMatchName
}

func (f functionStringWildcardAllMatch) Calculate(ctx *Context) (AttributeValue, error) {
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

			if !ok {
				return MakeBooleanValue(false), nil
			}
		}

		return MakeBooleanValue(true), nil

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

			if !ok {
				return MakeBooleanValue(false), nil
			}
		}

		return MakeBooleanValue(true), nil

	}

	return undefinedValue,
		bindError(
			bindError(newInvalidContainerArgType(pType, TypeSetOfStrings, TypeListOfStrings),
				"patterns",
			),
			f.describe(),
		)
}

type functionStringWildcardMatchAll struct {
	pattern Expression
	strs    Expression
}

func makeFunctionStringWildcardMatchAll(pattern, strs Expression) Expression {
	return functionStringWildcardMatchAll{
		pattern: pattern,
		strs:    strs}
}

func makeFunctionStringWildcardMatchAllAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function %q for Strings needs exactly two arguments but got %d",
			functionStringWildcardAllMatchName, len(args)))
	}

	return makeFunctionStringWildcardMatchAll(args[0], args[1])
}

func (f functionStringWildcardMatchAll) GetResultType() int {
	return TypeBoolean
}

func (f functionStringWildcardMatchAll) describe() string {
	return functionStringWildcardAllMatchName
}

func (f functionStringWildcardMatchAll) Calculate(ctx *Context) (AttributeValue, error) {
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

			if !ok {
				return MakeBooleanValue(false), nil
			}
		}

		return MakeBooleanValue(true), nil

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

			if !ok {
				return MakeBooleanValue(false), nil
			}
		}

		return MakeBooleanValue(true), nil

	}

	return undefinedValue,
		bindError(
			bindError(newInvalidContainerArgType(sType, TypeSetOfStrings, TypeListOfStrings),
				"strings",
			),
			f.describe(),
		)
}

func functionStringWildcardAllMatchValidator(args []Expression) functionMaker {
	if len(args) != 2 {
		return nil
	}

	pType := args[0].GetResultType()
	sType := args[1].GetResultType()

	if (pType == TypeSetOfStrings || pType == TypeListOfStrings) && sType == TypeString {
		return makeFunctionStringWildcardAllMatchAlt
	}

	if pType == TypeString && (sType == TypeSetOfStrings || sType == TypeListOfStrings) {
		return makeFunctionStringWildcardMatchAllAlt
	}

	return nil
}
