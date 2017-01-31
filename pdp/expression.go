package pdp

type ExpressionType interface {
	getResultType() int
	calculate(ctx *Context) (AttributeValueType, error)
}

type twoArgumentsFunctionType func(first ExpressionType, second ExpressionType) ExpressionType
type anyArgumentsFunctionType func(args []ExpressionType) ExpressionType

type argumentChecker func(args []ExpressionType) anyArgumentsFunctionType

var expressionArgumentCheckers map[string][]argumentChecker = map[string][]argumentChecker{
	"equal": {checkerFunctionStringEqual},
	"contains": {
		checkerFunctionStringContains,
		checkerFunctionNetworkContainsAddress,
		checkerFunctionSetOfStringContains,
		checkerFunctionSetOfNetworksContainsAddress,
		checkerFunctionSetOfDomainsContains},
	"not": {checkerFunctionBooleanNot},
	"or":  {checkerFunctionBooleanOr},
	"and": {checkerFunctionBooleanAnd}}
