package pdp

type ExpressionType interface {
	describe() string
	getResultType() int
	calculate(ctx *Context) (AttributeValueType, error)
}
