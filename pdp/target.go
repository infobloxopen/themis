package pdp

type Match struct {
	m Expression
}

type AllOf struct {
	m []Match
}

type AnyOf struct {
	a []AllOf
}

type Target struct {
	a []AnyOf
}

func MakeMatch(e Expression) Match {
	return Match{m: e}
}

func (m Match) describe() string {
	return "match"
}

func (m Match) calculate(ctx *Context) (bool, error) {
	v, err := ctx.calculateBooleanExpression(m.m)
	if err != nil {
		return false, bindError(err, m.describe())
	}

	return v, nil
}

func MakeAllOf() AllOf {
	return AllOf{m: []Match{}}
}

func (a AllOf) describe() string {
	return "all"
}

func (a AllOf) calculate(ctx *Context) (bool, error) {
	for _, e := range a.m {
		v, err := e.calculate(ctx)
		if err != nil {
			return true, bindError(err, a.describe())
		}

		if !v {
			return false, nil
		}
	}

	return true, nil
}

func (a *AllOf) Append(item Match) {
	a.m = append(a.m, item)
}

func MakeAnyOf() AnyOf {
	return AnyOf{a: []AllOf{}}
}

func (a AnyOf) describe() string {
	return "any"
}

func (a AnyOf) calculate(ctx *Context) (bool, error) {
	for _, e := range a.a {
		v, err := e.calculate(ctx)
		if err != nil {
			return false, bindError(err, a.describe())
		}

		if v {
			return true, nil
		}
	}

	return false, nil
}

func (a *AnyOf) Append(item AllOf) {
	a.a = append(a.a, item)
}

func MakeTarget() Target {
	return Target{a: []AnyOf{}}
}

func (t Target) describe() string {
	return "target"
}

func (t Target) calculate(ctx *Context) (bool, boundError) {
	for _, e := range t.a {
		v, err := e.calculate(ctx)
		if err != nil {
			return true, bindError(err, t.describe())
		}

		if !v {
			return false, nil
		}
	}

	return true, nil
}

func (t *Target) Append(item AnyOf) {
	t.a = append(t.a, item)
}

func makeMatchStatus(err boundError, effect int) Response {
	if effect == EffectDeny {
		return Response{EffectIndeterminateD, err, nil}
	}

	return Response{EffectIndeterminateP, err, nil}
}

func combineEffectAndStatus(err boundError, r Response) Response {
	if r.status != nil {
		err = newMultiError([]error{err, r.status})
	}

	if r.Effect == EffectNotApplicable {
		return Response{EffectNotApplicable, err, nil}
	}

	if r.Effect == EffectDeny || r.Effect == EffectIndeterminateD {
		return Response{EffectIndeterminateD, err, nil}
	}

	if r.Effect == EffectPermit || r.Effect == EffectIndeterminateP {
		return Response{EffectIndeterminateP, err, nil}
	}

	return Response{EffectIndeterminateDP, err, nil}
}

type twoArgumentsFunctionType func(first, second Expression) Expression

var TargetCompatibleExpressions = map[string]map[int]map[int]twoArgumentsFunctionType{
	"equal": {
		TypeString: {
			TypeString: makeFunctionStringEqual}},
	"contains": {
		TypeString: {
			TypeString: makeFunctionStringContains},
		TypeNetwork: {
			TypeAddress: makeFunctionNetworkContainsAddress},
		TypeSetOfStrings: {
			TypeString: makeFunctionSetOfStringsContains},
		TypeSetOfNetworks: {
			TypeAddress: makeFunctionSetOfNetworksContainsAddress},
		TypeSetOfDomains: {
			TypeDomain: makeFunctionSetOfDomainsContains}}}
