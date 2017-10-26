package jast

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

type (
	policyCombiningAlgParamBuilder func(ctx context, alg *caParams, policies []pdp.Evaluable) (interface{}, boundError)
	ruleCombiningAlgParamBuilder   func(ctx context, alg *caParams, rules []*pdp.Rule) (interface{}, boundError)
)

var (
	policyCombiningAlgParamBuilders = map[string]policyCombiningAlgParamBuilder{}
	ruleCombiningAlgParamBuilders   = map[string]ruleCombiningAlgParamBuilder{}
)

func init() {
	policyCombiningAlgParamBuilders["mapper"] = buildMapperPolicyCombiningAlgParams
	ruleCombiningAlgParamBuilders["mapper"] = buildMapperRuleCombiningAlgParams
}

type caParams struct {
	id     string
	defOk  bool
	defID  string
	errOk  bool
	errID  string
	arg    pdp.Expression
	subAlg interface{}
}

func checkPolicyID(ID string, policies []pdp.Evaluable) bool {
	for _, p := range policies {
		if pid, ok := p.GetID(); ok && ID == pid {
			return true
		}
	}

	return false
}

func buildMapperPolicyCombiningAlgParams(ctx context, alg *caParams, policies []pdp.Evaluable) (interface{}, boundError) {
	if alg.defOk {
		if !checkPolicyID(alg.defID, policies) {
			return nil, newMissingDefaultPolicyPCAError(alg.defID)
		}
	}

	if alg.errOk {
		if !checkPolicyID(alg.errID, policies) {
			return nil, newMissingErrorPolicyPCAError(alg.errID)
		}
	}

	var subAlg pdp.PolicyCombiningAlg
	if alg.subAlg != nil {
		maker, params, err := ctx.buildPolicyCombiningAlg(alg.subAlg, policies)
		if err != nil {
			return nil, err
		}
		subAlg = maker(nil, params)
	}

	return pdp.MapperPCAParams{
		Argument:  alg.arg,
		DefOk:     alg.defOk,
		Def:       alg.defID,
		ErrOk:     alg.errOk,
		Err:       alg.errID,
		Algorithm: subAlg}, nil
}

func checkRuleID(ID string, rules []*pdp.Rule) bool {
	for _, r := range rules {
		if rid, ok := r.GetID(); ok && ID == rid {
			return true
		}
	}

	return false
}

func buildMapperRuleCombiningAlgParams(ctx context, alg *caParams, rules []*pdp.Rule) (interface{}, boundError) {
	if alg.defOk {
		if !checkRuleID(alg.defID, rules) {
			return nil, newMissingDefaultRuleRCAError(alg.defID)
		}
	}

	if alg.errOk {
		if !checkRuleID(alg.errID, rules) {
			return nil, newMissingErrorRuleRCAError(alg.errID)
		}
	}

	var subAlg pdp.RuleCombiningAlg
	if alg.subAlg != nil {
		maker, params, err := ctx.buildRuleCombiningAlg(alg.subAlg, rules)
		if err != nil {
			return nil, err
		}
		subAlg = maker(nil, params)
	}

	return pdp.MapperRCAParams{
		Argument:  alg.arg,
		DefOk:     alg.defOk,
		Def:       alg.defID,
		ErrOk:     alg.errOk,
		Err:       alg.errID,
		Algorithm: subAlg}, nil
}

func (ctx context) buildRuleCombiningAlg(alg interface{}, rules []*pdp.Rule) (pdp.RuleCombiningAlgMaker, interface{}, boundError) {
	switch alg := alg.(type) {
	case *caParams:
		id := strings.ToLower(alg.id)
		maker, ok := pdp.RuleCombiningParamAlgs[id]
		if !ok {
			return nil, nil, newUnknownRCAError(alg.id)
		}

		paramBuilder, ok := ruleCombiningAlgParamBuilders[id]
		if !ok {
			return nil, nil, newNotImplementedRCAError(alg.id)
		}

		params, err := paramBuilder(ctx, alg, rules)
		if err != nil {
			return nil, nil, bindError(err, alg.id)
		}

		return maker, params, nil
	case string:
		maker, ok := pdp.RuleCombiningAlgs[strings.ToLower(alg)]
		if !ok {
			return nil, nil, newUnknownRCAError(alg)
		}

		return maker, nil, nil
	}

	return nil, nil, newInvalidRCAError(alg)
}

func (ctx context) buildPolicyCombiningAlg(alg interface{}, policies []pdp.Evaluable) (pdp.PolicyCombiningAlgMaker, interface{}, boundError) {
	switch alg := alg.(type) {
	case *caParams:
		id := strings.ToLower(alg.id)
		maker, ok := pdp.PolicyCombiningParamAlgs[id]
		if !ok {
			return nil, nil, newUnknownPCAError(alg.id)
		}

		paramBuilder, ok := policyCombiningAlgParamBuilders[id]
		if !ok {
			return nil, nil, newNotImplementedPCAError(alg.id)
		}

		params, err := paramBuilder(ctx, alg, policies)
		if err != nil {
			return nil, nil, bindError(err, alg.id)
		}

		return maker, params, nil
	case string:
		maker, ok := pdp.PolicyCombiningAlgs[strings.ToLower(alg)]
		if !ok {
			return nil, nil, newUnknownPCAError(alg)
		}

		return maker, nil, nil
	}

	return nil, nil, newInvalidPCAError(alg)
}

func (ctx context) unmarshalCombiningAlgObj(d *json.Decoder) (*caParams, error) {
	var (
		mapOk  bool
		params caParams
		id     string
	)

	if err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		var err error

		switch strings.ToLower(k) {
		case yastTagID:
			id, err = jparser.GetString(d, "algorithm id")
			return err

		case yastTagMap:
			mapOk = true
			err = jparser.CheckObjectStart(d, "expression")
			if err != nil {
				return bindErrorf(err, "%s", id)
			}

			params.arg, err = ctx.unmarshalExpression(d)
			if err != nil {
				return bindErrorf(err, "%s", id)
			}

			t := params.arg.GetResultType()
			if t != pdp.TypeString && t != pdp.TypeSetOfStrings && t != pdp.TypeListOfStrings {
				return bindErrorf(newMapperArgumentTypeError(t), "%s", id)
			}

			return nil

		case yastTagDefault:
			params.defOk = true
			params.defID, err = jparser.GetString(d, "algorithm default id")
			if err != nil {
				return bindErrorf(err, "%s", id)
			}

			return nil

		case yastTagError:
			params.errOk = true
			params.errID, err = jparser.GetString(d, "algorithm error id")
			if err != nil {
				return bindErrorf(err, "%s", id)
			}

			return nil

		case yastTagAlg:
			params.subAlg, err = ctx.unmarshalCombiningAlg(d)
			if err != nil {
				return bindErrorf(err, "%s", id)
			}

			return nil
		}

		return newUnknownAttributeError(k)
	}, "algorithm"); err != nil {
		return nil, err
	}

	if id == "" {
		return nil, newMissingAttributeError(yastTagID, "algorithm")
	}

	if !mapOk {
		return nil, newMissingAttributeError(yastTagMap, "algorithm")
	}

	params.id = id

	return &params, nil
}

func (ctx context) unmarshalCombiningAlg(d *json.Decoder) (interface{}, error) {
	t, err := d.Token()
	if err != nil {
		return nil, err
	}

	switch t := t.(type) {
	case json.Delim:
		if t.String() == jparser.DelimObjectStart {
			return ctx.unmarshalCombiningAlgObj(d)
		}

		return nil, newParseCAError(t)
	case string:
		return t, nil
	default:
		return nil, newParseCAError(t)
	}
}
