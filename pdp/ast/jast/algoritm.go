package jast

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

type (
	policyCombiningAlgParamUnmarshaler func(ctx context, m map[interface{}]interface{}, policies []pdp.Evaluable) (interface{}, boundError)
	ruleCombiningAlgParamUnmarshaler   func(ctx context, m map[interface{}]interface{}, rules []*pdp.Rule) (interface{}, boundError)
)

var (
	policyCombiningAlgParamUnmarshalers = map[string]policyCombiningAlgParamUnmarshaler{}
	ruleCombiningAlgParamUnmarshalers   = map[string]ruleCombiningAlgParamUnmarshaler{}
)

func init() {
	policyCombiningAlgParamUnmarshalers["mapper"] = unmarshalMapperPolicyCombiningAlgParams
	ruleCombiningAlgParamUnmarshalers["mapper"] = unmarshalMapperRuleCombiningAlgParams
}

func checkPolicyID(ID string, policies []pdp.Evaluable) bool {
	for _, p := range policies {
		if pid, ok := p.GetID(); ok && ID == pid {
			return true
		}
	}

	return false
}

func unmarshalMapperPolicyCombiningAlgParams(ctx context, m map[interface{}]interface{}, policies []pdp.Evaluable) (interface{}, boundError) {
	v, ok := m[yastTagMap]
	if !ok {
		return nil, newMissingMapPCAParamError()
	}

	arg, err := ctx.unmarshalExpression(v)
	if err != nil {
		return nil, err
	}

	t := arg.GetResultType()
	if t != pdp.TypeString && t != pdp.TypeSetOfStrings && t != pdp.TypeListOfStrings {
		return nil, newMapperArgumentTypeError(t)
	}

	defID, defOk, err := ctx.extractStringOpt(m, yastTagDefault, "default policy id")
	if err != nil {
		return nil, err
	}

	if defOk {
		if !checkPolicyID(defID, policies) {
			return nil, newMissingDefaultPolicyPCAError(defID)
		}
	}

	errID, errOk, err := ctx.extractStringOpt(m, yastTagError, "on error policy id")
	if err != nil {
		return nil, err
	}

	if errOk {
		if !checkPolicyID(errID, policies) {
			return nil, newMissingErrorPolicyPCAError(errID)
		}
	}

	var subAlg pdp.PolicyCombiningAlg
	if t == pdp.TypeSetOfStrings || t == pdp.TypeListOfStrings {
		maker, params, err := ctx.unmarshalPolicyCombiningAlg(m, nil)
		if err != nil {
			return nil, err
		}
		subAlg = maker(nil, params)
	}

	return pdp.MapperPCAParams{
		Argument:  arg,
		DefOk:     defOk,
		Def:       defID,
		ErrOk:     errOk,
		Err:       errID,
		Algorithm: subAlg}, nil
}

func (ctx context) unmarshalPolicyCombiningAlgObj(m map[interface{}]interface{}, policies []pdp.Evaluable) (pdp.PolicyCombiningAlgMaker, interface{}, boundError) {
	ID, err := ctx.extractString(m, yastTagID, "algorithm id")
	if err != nil {
		return nil, nil, err
	}

	s := strings.ToLower(ID)
	maker, ok := pdp.PolicyCombiningParamAlgs[s]
	if !ok {
		return nil, nil, newUnknownPCAError(ID)
	}

	paramUnmarshaler, ok := policyCombiningAlgParamUnmarshalers[s]
	if !ok {
		return nil, nil, newNotImplementedPCAError(ID)
	}

	params, err := paramUnmarshaler(ctx, m, policies)
	if err != nil {
		return nil, nil, bindError(err, ID)
	}

	return maker, params, nil
}

func (ctx context) unmarshalPolicyCombiningAlg(m map[interface{}]interface{}, policies []pdp.Evaluable) (pdp.PolicyCombiningAlgMaker, interface{}, boundError) {
	v, ok := m[yastTagAlg]
	if !ok {
		return nil, nil, newMissingPCAError()
	}

	switch alg := v.(type) {
	case string:
		maker, ok := pdp.PolicyCombiningAlgs[strings.ToLower(alg)]
		if !ok {
			return nil, nil, newUnknownPCAError(alg)
		}

		return maker, nil, nil

	case map[interface{}]interface{}:
		return ctx.unmarshalPolicyCombiningAlgObj(alg, policies)
	}

	return nil, nil, newInvalidPCAError(v)
}

func checkRuleID(ID string, rules []*pdp.Rule) bool {
	for _, r := range rules {
		if rid, ok := r.GetID(); ok && ID == rid {
			return true
		}
	}

	return false
}

func unmarshalMapperRuleCombiningAlgParams(ctx context, m map[interface{}]interface{}, rules []*pdp.Rule) (interface{}, boundError) {
	v, ok := m[yastTagMap]
	if !ok {
		return nil, newMissingMapRCAParamError()
	}

	arg, err := ctx.unmarshalExpression(v)
	if err != nil {
		return nil, err
	}

	t := arg.GetResultType()
	if t != pdp.TypeString && t != pdp.TypeSetOfStrings && t != pdp.TypeListOfStrings {
		return nil, newMapperArgumentTypeError(t)
	}

	defID, defOk, err := ctx.extractStringOpt(m, yastTagDefault, "default rule id")
	if err != nil {
		return nil, err
	}

	if defOk {
		if !checkRuleID(defID, rules) {
			return nil, newMissingDefaultRuleRCAError(defID)
		}
	}

	errID, errOk, err := ctx.extractStringOpt(m, yastTagError, "on error rule id")
	if err != nil {
		return nil, err
	}

	if errOk {
		if !checkRuleID(errID, rules) {
			return nil, newMissingErrorRuleRCAError(errID)
		}
	}

	var subAlg pdp.RuleCombiningAlg
	if t == pdp.TypeSetOfStrings || t == pdp.TypeListOfStrings {
		maker, params, err := ctx.unmarshalRuleCombiningAlg(m, nil)
		if err != nil {
			return nil, err
		}
		subAlg = maker(nil, params)
	}

	return pdp.MapperRCAParams{
		Argument:  arg,
		DefOk:     defOk,
		Def:       defID,
		ErrOk:     errOk,
		Err:       errID,
		Algorithm: subAlg}, nil
}

func (ctx context) unmarshalRuleCombiningAlgObj(m map[interface{}]interface{}, rules []*pdp.Rule) (pdp.RuleCombiningAlgMaker, interface{}, boundError) {
	ID, err := ctx.extractString(m, yastTagID, "algorithm id")
	if err != nil {
		return nil, nil, err
	}

	s := strings.ToLower(ID)
	maker, ok := pdp.RuleCombiningParamAlgs[s]
	if !ok {
		return nil, nil, newUnknownRCAError(ID)
	}

	paramUnmarshaler, ok := ruleCombiningAlgParamUnmarshalers[s]
	if !ok {
		return nil, nil, newNotImplementedRCAError(ID)
	}

	params, err := paramUnmarshaler(ctx, m, rules)
	if err != nil {
		return nil, nil, bindError(err, ID)
	}

	return maker, params, nil
}

func (ctx context) unmarshalRuleCombiningAlg(m map[interface{}]interface{}, rules []*pdp.Rule) (pdp.RuleCombiningAlgMaker, interface{}, boundError) {
	v, ok := m[yastTagAlg]
	if !ok {
		return nil, nil, newMissingRCAError()
	}

	switch alg := v.(type) {
	case string:
		maker, ok := pdp.RuleCombiningAlgs[strings.ToLower(alg)]
		if !ok {
			return nil, nil, newUnknownRCAError(alg)
		}

		return maker, nil, nil

	case map[interface{}]interface{}:
		return ctx.unmarshalRuleCombiningAlgObj(alg, rules)
	}

	return nil, nil, newInvalidRCAError(v)
}

func (ctx *context) decodeCombiningAlg(d *json.Decoder) (map[interface{}]interface{}, error) {
	v, err := ctx.decodeUndefined(d, "algorithm")
	if err != nil {
		return nil, err
	}

	return map[interface{}]interface{}{yastTagAlg: v}, nil
}
