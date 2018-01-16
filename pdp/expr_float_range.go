package pdp

import "fmt"

type functionFloatRange struct {
	min Expression
	max Expression
	val Expression
}

func makeFunctionFloatRange(min, max, val Expression) Expression {
	return functionFloatRange{
		min: min,
		max: max,
		val: val,
	}
}

func makeFunctionFloatRangeAlt(args []Expression) Expression {
	if len(args) != 3 {
		panic(fmt.Errorf("function \"Range\" for Float needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionFloatRange(args[0], args[1], args[2])
}

func (f functionFloatRange) GetResultType() int {
	return TypeString
}

func (f functionFloatRange) describe() string {
	return "range"
}

func (f functionFloatRange) calculate(ctx *Context) (AttributeValue, error) {
	min, err := ctx.calculateFloatExpression(f.min)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "min argument"), f.describe())
	}

	max, err := ctx.calculateFloatExpression(f.max)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "max argument"), f.describe())
	}

	val, err := ctx.calculateFloatExpression(f.val)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "val argument"), f.describe())
	}

	switch {
	case val < min:
		return MakeStringValue("Below"), nil
	case max < val:
		return MakeStringValue("Above"), nil
	}
	return MakeStringValue("Within"), nil
}

func functionFloatRangeValidator(args []Expression) functionMaker {
	if len(args) != 3 || args[0].GetResultType() != TypeFloat || args[1].GetResultType() != TypeFloat || args[2].GetResultType() != TypeFloat {
		return nil
	}

	return makeFunctionFloatRangeAlt
}
