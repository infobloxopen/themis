package jast

import (
	"encoding/json"
	"strings"

	"github.com/infobloxopen/themis/jparser"
	"github.com/infobloxopen/themis/pdp"
)

func (ctx *context) decodeEntity(d *json.Decoder) (interface{}, error) {
	if err := jparser.CheckObjectStart(d, "entity"); err != nil {
		return nil, err
	}

	var (
		hidden      bool = true
		isPolicy    bool
		isPolicySet bool
		isRule      bool

		id       string
		effect   int
		policies []pdp.Evaluable
		rules    []*pdp.Rule
		target   pdp.Target
		cond     pdp.Expression
		obligs   []pdp.AttributeAssignmentExpression

		algObj map[interface{}]interface{}
	)

	err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		var err error

		switch strings.ToLower(k) {
		case yastTagID:
			hidden = false
			id, err = jparser.GetString(d, "policy or set or rule id")
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
				return bindError(err, makeSource("policy set", id, hidden))
			}
			return nil

		case yastTagRules:
			isPolicy = true
			rules, err = ctx.decodeRules(d)
			if err != nil {
				return bindError(err, makeSource("policy", id, hidden))
			}
			return nil

		case yastTagEffect:
			isRule = true
			var s string
			src := makeSource("rule", id, hidden)
			s, err = jparser.GetString(d, "effect")
			if err != nil {
				return bindError(err, src)
			}

			var ok bool
			effect, ok = pdp.EffectIDs[strings.ToLower(s)]
			if !ok {
				return bindError(newUnknownEffectError(s), src)
			}
			return nil

		case yastTagCondition:
			var v interface{}
			v, err = ctx.decodeUndefined(d, "condition")
			if err != nil {
				return err
			}

			m := map[interface{}]interface{}{yastTagCondition: v}
			cond, err = ctx.unmarshalCondition(m)

			return err
		}

		return newUnknownAttributeError(k)
	}, "entity")
	if err != nil {
		return nil, err
	}

	if isRule && isPolicy || isRule && isPolicySet || isPolicy && isPolicySet {
		tags := []string{}
		if isPolicy {
			tags = append(tags, yastTagRules)
		}

		if isPolicySet {
			tags = append(tags, yastTagPolicies)
		}

		if isRule {
			tags = append(tags, yastTagEffect)
		}

		return nil, newEntityAmbiguityError(tags)
	}

	if isPolicySet {
		alg, params, err := ctx.unmarshalPolicyCombiningAlg(algObj, policies)
		if err != nil {
			return nil, bindError(err, makeSource("policy set", id, hidden))
		}

		return pdp.NewPolicySet(id, hidden, target, policies, alg, params, obligs), nil
	}

	if isPolicy {
		alg, params, err := ctx.unmarshalRuleCombiningAlg(algObj, rules)
		if err != nil {
			return nil, bindError(err, makeSource("policy", id, hidden))
		}

		return pdp.NewPolicy(id, hidden, target, rules, alg, params, obligs), nil
	}

	if isRule {
		return pdp.NewRule(id, hidden, target, cond, effect, obligs), nil
	}

	return nil, newEntityMissingKeyError()
}

func (ctx *context) decodeCommand(d *json.Decoder, u *pdp.PolicyUpdate) error {
	var (
		op     int
		path   []string
		entity interface{}
	)

	err := jparser.UnmarshalObject(d, func(k string, d *json.Decoder) error {
		var err error

		switch strings.ToLower(k) {
		case yastTagOp:
			var s string
			s, err = jparser.GetString(d, "operation")
			if err != nil {
				return err
			}

			var ok bool
			op, ok = pdp.UpdateOpIDs[strings.ToLower(s)]
			if !ok {
				return newUnknownPolicyUpdateOperationError(s)
			}

			return nil

		case yastTagPath:
			path = []string{}
			err = jparser.GetStringSequence(d, "path", func(idx int, s string) error {
				path = append(path, s)
				return nil
			})

			return err

		case yastTagEntity:
			if op == pdp.UOAdd {
				entity, err = ctx.decodeEntity(d)
			}

			return err
		}

		return newUnknownAttributeError(k)
	}, "command")
	if err != nil {
		return err
	}

	u.Append(op, path, entity)

	return nil
}

func (ctx *context) decodeCommands(d *json.Decoder, u *pdp.PolicyUpdate) error {
	if err := jparser.CheckArrayStart(d, "commands"); err != nil {
		return err
	}

	if err := jparser.UnmarshalObjectArray(d, func(idx int, d *json.Decoder) error {
		if err := ctx.decodeCommand(d, u); err != nil {
			return bindErrorf(err, "%d", idx)
		}

		return nil
	}, "commands"); err != nil {
		return err
	}

	return nil
}
