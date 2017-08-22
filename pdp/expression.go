package pdp

type Expression interface {
	GetResultType() int
	calculate(ctx *Context) (AttributeValue, error)
}

type functionMaker func(args []Expression) Expression
type functionArgumentValidator func(args []Expression) functionMaker

var FunctionArgumentValidators = map[string][]functionArgumentValidator{
	"equal": {functionStringEqualValidator},
	"contains": {
		functionStringContainsValidator,
		functionNetworkContainsAddressValidator,
		functionSetOfStringsContainsValidator,
		functionSetOfNetworksContainsAddressValidator,
		functionSetOfDomainsContainsValidator},
	"not": {functionBooleanNotValidator},
	"or":  {functionBooleanOrValidator},
	"and": {functionBooleanAndValidator}}
