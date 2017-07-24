package pdp

type expression interface {
	calculate(ctx *Context) (attributeValue, error)
}
