package jast

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

func makeSource(desc string, id string, hidden bool) string {
	if hidden {
		return fmt.Sprintf("hidden %s", desc)
	}
	return fmt.Sprintf("%s \"%s\"", desc, id)
}

func (ctx *context) decodePolicies(d *json.Decoder) ([]pdp.Evaluable, error) {
	if err := jparser.CheckArrayStart(d, "policy set"); err != nil {
		return nil, err
	}

	policies := []pdp.Evaluable{}
	if err := jparser.UnmarshalObjectArray(d, func(idx int, d *json.Decoder) error {
		e, err := ctx.decodeEvaluable(d)
		if err != nil {
			return bindErrorf(err, "%d", idx)
		}

		policies = append(policies, e)

		return nil
	}, "policy set"); err != nil {
		return nil, err
	}

	return policies, nil
}

func (ctx *context) decodeEvaluable(d *json.Decoder) (pdp.Evaluable, error) {
	var (
		hidden      bool = true
		isPolicy    bool
		isPolicySet bool

		pid      string
		policies []pdp.Evaluable
		rules    []*pdp.Rule
		target   pdp.Target
		obligs   []pdp.AttributeAssignmentExpression

		algObj map[interface{}]interface{}
	)

	err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		var err error

		switch strings.ToLower(k) {
		case yastTagID:
			hidden = false
			pid, err = jparser.GetString(d, "policy or policy set id")
			return err

		case yastTagAlg:
			algObj, err = ctx.decodeCombiningAlg(d)
			return err

		case yastTagTarget:
			target, err = ctx.decodeTarget(d)
			return err

		case yastTagObligation:
			obligs, err = ctx.decodeObligations(d)
			return err

		case yastTagPolicies:
			isPolicySet = true
			policies, err = ctx.decodePolicies(d)
			if err != nil {
				return bindError(err, makeSource("policy set", pid, hidden))
			}
			return nil

		case yastTagRules:
			isPolicy = true
			rules, err = ctx.decodeRules(d)
			if err != nil {
				return bindError(err, makeSource("policy", pid, hidden))
			}
			return nil
		}

		return newUnknownAttributeError(k)
	}, "policy or policy set")
	if err != nil {
		return nil, err
	}

	if isPolicy && isPolicySet {
		return nil, newPolicyAmbiguityError()
	}

	if isPolicySet {
		alg, params, err := ctx.unmarshalPolicyCombiningAlg(algObj, policies)
		if err != nil {
			return nil, bindError(err, makeSource("policy set", pid, hidden))
		}

		return pdp.NewPolicySet(pid, hidden, target, policies, alg, params, obligs), nil
	}

	if isPolicy {
		alg, params, err := ctx.unmarshalRuleCombiningAlg(algObj, rules)
		if err != nil {
			return nil, bindError(err, makeSource("policy", pid, hidden))
		}

		return pdp.NewPolicy(pid, hidden, target, rules, alg, params, obligs), nil
	}

	return nil, newPolicyMissingKeyError()
}

func (ctx *context) decodeRootPolicy(d *json.Decoder) error {
	if err := jparser.CheckObjectStart(d, "root policy or policy set"); err != nil {
		return err
	}

	e, err := ctx.decodeEvaluable(d)
	if err != nil {
		return err
	}

	ctx.rootPolicy = e
	return nil
}
