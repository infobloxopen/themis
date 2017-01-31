package pdp

type ExpressionType interface {
	getResultType() int
	calculate(ctx *Context) (AttributeValueType, error)
}
