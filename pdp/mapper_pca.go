package pdp

import "fmt"

type MapperPCAParams struct {
	Argument      ExpressionType
	PoliciesMap   map[string]EvaluableType
	DefaultPolicy EvaluableType
	ErrorPolicy   EvaluableType
	SubAlg        PolicyCombiningAlgType
	AlgParams     interface{}
}

func calculateErrorPolicy(policy EvaluableType, ctx *Context, err error) ResponseType {
	if policy != nil {
		return policy.Calculate(ctx)
	}

	return ResponseType{EffectIndeterminate, fmt.Sprintf("Mapper Policy Combining Algorithm: %s", err), nil}
}

func getPoliciesMap(policies []EvaluableType, params *MapperPCAParams) map[string]EvaluableType {
	if params.PoliciesMap != nil {
		return params.PoliciesMap
	}

	m := make(map[string]EvaluableType)
	for _, policy := range policies {
		m[policy.getID()] = policy
	}

	return m
}

func collectSubPolicies(IDs map[string]bool, m map[string]EvaluableType) []EvaluableType {
	policies := []EvaluableType{}
	for ID := range IDs {
		policy, ok := m[ID]
		if ok {
			policies = append(policies, policy)
		}
	}

	return policies
}

func MapperPCA(policies []EvaluableType, params interface{}, ctx *Context) ResponseType {
	mapperParams := params.(MapperPCAParams)

	v, err := mapperParams.Argument.calculate(ctx)
	if err != nil {
		switch err.(type) {
		case MissingValueError, *MissingValueError:
			if mapperParams.DefaultPolicy != nil {
				return mapperParams.DefaultPolicy.Calculate(ctx)
			}
		}

		return calculateErrorPolicy(mapperParams.ErrorPolicy, ctx, err)
	}

	if mapperParams.SubAlg != nil {
		IDs, err := getSetOfIDs(v)
		if err != nil {
			return calculateErrorPolicy(mapperParams.ErrorPolicy, ctx, err)
		}

		return mapperParams.SubAlg(collectSubPolicies(IDs, getPoliciesMap(policies, &mapperParams)),
			mapperParams.AlgParams, ctx)
	}

	ID, err := ExtractStringValue(v, "argument")
	if err != nil {
		return calculateErrorPolicy(mapperParams.ErrorPolicy, ctx, err)
	}

	if mapperParams.PoliciesMap != nil {
		policy, ok := mapperParams.PoliciesMap[ID]
		if ok {
			return policy.Calculate(ctx)
		}
	} else {
		for _, policy := range policies {
			if policy.getID() == ID {
				return policy.Calculate(ctx)
			}
		}
	}

	if mapperParams.DefaultPolicy != nil {
		return mapperParams.DefaultPolicy.Calculate(ctx)
	}

	return ResponseType{EffectNotApplicable, "Ok", nil}
}
