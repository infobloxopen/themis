package pdp

import (
	"fmt"

	"github.com/infobloxopen/go-trees/strtree"
)

type mapperPCA struct {
	argument  Expression
	policies  *strtree.Tree
	def       Evaluable
	err       Evaluable
	algorithm PolicyCombiningAlg
}

// MapperPCAParams gathers all parameters of mapper policy combining algorithm.
type MapperPCAParams struct {
	// Argument represent expression which value is used to get nested policy
	// set or policy (or list of them).
	Argument Expression

	// DefOk indicates if Def contains valid value.
	DefOk bool
	// Def contains id of default policy set or policy (the default policy
	// is used when Argument expression evaluates to a value which doesn't
	// match to any id). This value is used only if DefOk is true.
	Def string

	// ErrOk indicateis if Err contains valid value.
	ErrOk bool
	// Err ontains id of policy set or policy to use in case of error (when
	// Argument can't be evaluated).
	Err string

	// Algorithm is additional policy combining algorithm which is used when
	// argument can return several ids.
	Algorithm PolicyCombiningAlg
}

func collectSubPolicies(IDs []string, m *strtree.Tree) []Evaluable {
	policies := []Evaluable{}
	for _, ID := range IDs {
		policy, ok := m.Get(ID)
		if ok {
			policies = append(policies, policy.(Evaluable))
		}
	}

	return policies
}

func makeMapperPCA(policies []Evaluable, params interface{}) PolicyCombiningAlg {
	mapperParams, ok := params.(MapperPCAParams)
	if !ok {
		panic(fmt.Errorf("Mapper policy combining algorithm maker expected MapperPCAParams structure as params "+
			"but got %T", params))
	}

	var (
		m   *strtree.Tree
		def Evaluable
		err Evaluable
	)

	if policies != nil {
		m = strtree.NewTree()
		count := 0
		for _, p := range policies {
			if pid, ok := p.GetID(); ok {
				m.InplaceInsert(pid, p)
				count++
			}
		}

		if count > 0 {
			if mapperParams.DefOk {
				if v, ok := m.Get(mapperParams.Def); ok {
					def = v.(Evaluable)
				}
			}

			if mapperParams.ErrOk {
				if v, ok := m.Get(mapperParams.Err); ok {
					err = v.(Evaluable)
				}
			}
		} else {
			m = nil
		}
	}

	return mapperPCA{
		argument:  mapperParams.Argument,
		policies:  m,
		def:       def,
		err:       err,
		algorithm: mapperParams.Algorithm}
}

func (a mapperPCA) describe() string {
	return "mapper"
}

func (a mapperPCA) calculateErrorPolicy(ctx *Context, err error) Response {
	if a.err != nil {
		return a.err.Calculate(ctx)
	}

	return Response{EffectIndeterminate, bindError(err, a.describe()), nil}
}

func (a mapperPCA) getPoliciesMap(policies []Evaluable) *strtree.Tree {
	if a.policies != nil {
		return a.policies
	}

	r := strtree.NewTree()
	count := 0
	for _, p := range policies {
		if pid, ok := p.GetID(); ok {
			r.InplaceInsert(pid, p)
			count++
		}
	}

	if count > 0 {
		return r
	}

	return nil
}

func (a mapperPCA) add(ID string, child, old Evaluable) PolicyCombiningAlg {
	def := a.def
	if old != nil && old == def {
		def = child
	}

	err := a.err
	if old != nil && old == err {
		err = child
	}

	return mapperPCA{
		argument:  a.argument,
		policies:  a.policies.Insert(ID, child),
		def:       def,
		err:       err,
		algorithm: a.algorithm}
}

func (a mapperPCA) del(ID string, old Evaluable) PolicyCombiningAlg {
	def := a.def
	if old != nil && old == def {
		def = nil
	}

	err := a.err
	if old != nil && old == err {
		err = nil
	}

	policies := a.policies
	if policies != nil {
		policies, _ = a.policies.Delete(ID)
		if policies.IsEmpty() {
			policies = nil
		}
	}

	return mapperPCA{
		argument:  a.argument,
		policies:  policies,
		def:       def,
		err:       err,
		algorithm: a.algorithm}
}

func (a mapperPCA) execute(policies []Evaluable, ctx *Context) Response {
	v, err := a.argument.calculate(ctx)
	if err != nil {
		switch err.(type) {
		case *missingValueError:
			if a.def != nil {
				return a.def.Calculate(ctx)
			}
		}

		return a.calculateErrorPolicy(ctx, err)
	}

	if a.algorithm != nil {
		IDs, err := getSetOfIDs(v)
		if err != nil {
			return a.calculateErrorPolicy(ctx, err)
		}

		r := a.algorithm.execute(collectSubPolicies(IDs, a.getPoliciesMap(policies)), ctx)
		if r.Effect == EffectNotApplicable && a.def != nil {
			return a.def.Calculate(ctx)
		}

		return r
	}

	ID, err := v.str()
	if err != nil {
		return a.calculateErrorPolicy(ctx, err)
	}

	if a.policies != nil {
		policy, ok := a.policies.Get(ID)
		if ok {
			return policy.(Evaluable).Calculate(ctx)
		}
	} else {
		for _, policy := range policies {
			if PID, ok := policy.GetID(); ok && PID == ID {
				return policy.Calculate(ctx)
			}
		}
	}

	if a.def != nil {
		return a.def.Calculate(ctx)
	}

	return Response{EffectNotApplicable, nil, nil}
}
