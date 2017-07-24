package pdp

type match struct {
	m expression
}

type allOf struct {
	m []match
}

type anyOf struct {
	a []allOf
}

type target struct {
	a []anyOf
}

func (m match) describe() string {
	return "match"
}

func (m match) calculate(ctx *Context) (bool, error) {
	v, err := ctx.calculateBooleanExpression(m.m)
	if err != nil {
		return false, bindError(err, m.describe())
	}

	return v, nil
}

func (a allOf) describe() string {
	return "all"
}

func (a allOf) calculate(ctx *Context) (bool, error) {
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

func (a anyOf) describe() string {
	return "any"
}

func (a anyOf) calculate(ctx *Context) (bool, error) {
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

func (t target) describe() string {
	return "target"
}

func (t target) calculate(ctx *Context) (bool, boundError) {
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

func makeMatchStatus(err boundError, effect int) Response {
	if effect == EffectDeny {
		return Response{EffectIndeterminateD, err, nil}
	}

	return Response{EffectIndeterminateP, err, nil}
}

func combineEffectAndStatus(err boundError, r Response) Response {
	if r.status != nil {
		err = newNoSrcMultiError(err, r.status)
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
