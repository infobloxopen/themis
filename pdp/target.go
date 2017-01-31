package pdp

type MatchType struct {
	Match ExpressionType
}

type AllOfType struct {
	Matches []MatchType
}

type AnyOfType struct {
	AllOf []AllOfType
}

type TargetType struct {
	AnyOf []AnyOfType
}

var targetCompatibleExpressions map[string]map[int]map[int]twoArgumentsFunctionType = map[string]map[int]map[int]twoArgumentsFunctionType{
	"equal": map[int]map[int]twoArgumentsFunctionType{
		DataTypeString: map[int]twoArgumentsFunctionType{
			DataTypeString: makeFunctionStringEqual}},
	"contains": map[int]map[int]twoArgumentsFunctionType{
		DataTypeString: map[int]twoArgumentsFunctionType{
			DataTypeString: makeFunctionStringContains},
		DataTypeNetwork: map[int]twoArgumentsFunctionType{
			DataTypeAddress: makeFunctionNetworkContainsAddress},
		DataTypeSetOfStrings: map[int]twoArgumentsFunctionType{
			DataTypeString: makeFunctionSetOfStringContains},
		DataTypeSetOfNetworks: map[int]twoArgumentsFunctionType{
			DataTypeAddress: makeFunctionSetOfNetworksContainsAddress},
		DataTypeSetOfDomains: map[int]twoArgumentsFunctionType{
			DataTypeDomain: makeFunctionSetOfDomainsContains}}}

func (m MatchType) calculate(ctx *Context) (bool, error) {
	v, err := m.Match.calculate(ctx)
	if err != nil {
		return false, err
	}

	return ExtractBooleanValue(v, "a result of match expression")
}

func (a AllOfType) calculate(ctx *Context) (bool, error) {
	for _, e := range a.Matches {
		v, err := e.calculate(ctx)
		if err != nil {
			return true, err
		}

		if !v {
			return false, nil
		}
	}

	return true, nil
}

func (a AnyOfType) calculate(ctx *Context) (bool, error) {
	for _, e := range a.AllOf {
		v, err := e.calculate(ctx)
		if err != nil {
			return false, err
		}

		if v {
			return true, nil
		}
	}

	return false, nil
}

func (t TargetType) calculate(ctx *Context) (bool, error) {
	for _, e := range t.AnyOf {
		v, err := e.calculate(ctx)
		if err != nil {
			return true, err
		}

		if !v {
			return false, nil
		}
	}

	return true, nil
}

func MakeAll(m ...MatchType) AllOfType {
	return AllOfType{m}
}

func MakeAny(a ...AllOfType) AnyOfType {
	return AnyOfType{a}
}

func MakeTarget(a ...AnyOfType) TargetType {
	return TargetType{a}
}
