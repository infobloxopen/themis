package jast

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pdp/jcon"
)

func makeSource(desc string, ID string, hidden bool, idx int) string {
	src := fmt.Sprintf("hidden %s", desc)
	if !hidden {
		src = fmt.Sprintf("%s \"%s\"", desc, ID)
	}

	if idx > 0 {
		src = fmt.Sprintf("(%d) %s", idx, src)
	}

	return src
}

func (ctx *context) decodePolicies(d *json.Decoder) ([]pdp.Evaluable, error) {
	if err := jcon.CheckArrayStart(d, "policy set"); err != nil {
		return nil, err
	}

	policies := []pdp.Evaluable{}
	idx := 0

	err := jcon.UnmarshalObjectArray(d, func(d *json.Decoder) error {
		idx++

		e, err := ctx.decodeEvaluable(d, idx)
		if err != nil {
			return err
		}

		policies = append(policies, e)

		return nil
	}, "policy set")
	if err != nil {
		return nil, err
	}

	return policies, nil
}

func (ctx *context) decodeEvaluable(d *json.Decoder, i int) (pdp.Evaluable, error) {
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

	err := jcon.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		var err error

		switch strings.ToLower(k) {
		case yastTagID:
			hidden = false
			pid, err = jcon.GetString(d, "policy or policy set id")
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
			return err

		case yastTagRules:
			isPolicy = true
			rules, err = ctx.decodeRules(d)
			return err
		}

		return newUnknownAttributeError(k)
	}, "root")
	if err != nil {
		return nil, err
	}

	if isPolicy && isPolicySet {
		return nil, newPolicyAmbiguityError()
	}

	if isPolicySet {
		alg, params, err := ctx.unmarshalPolicyCombiningAlg(algObj, policies)
		if err != nil {
			return nil, err
		}

		return pdp.NewPolicySet(pid, hidden, target, policies, alg, params, obligs), nil
	}

	if isPolicy {
		alg, params, err := ctx.unmarshalRuleCombiningAlg(algObj, rules)
		if err != nil {
			return nil, err
		}

		return pdp.NewPolicy(pid, hidden, target, rules, alg, params, obligs), nil
	}

	return nil, newPolicyMissingKeyError()
}

func (ctx *context) decodeRootPolicy(d *json.Decoder) error {
	if err := jcon.CheckObjectStart(d, "root policy or policy set"); err != nil {
		return err
	}

	e, err := ctx.decodeEvaluable(d, 0)
	if err != nil {
		return err
	}

	ctx.rootPolicy = e
	return nil
}
