package yast

import (
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

func (ctx *context) unmarshalEntity(m map[interface{}]interface{}) (interface{}, error) {
	m, err := ctx.extractMap(m, yastTagEntity, "policy or set or rule")
	if err != nil {
		return nil, err
	}

	ID, ok, err := ctx.extractStringOpt(m, yastTagID, "policy or set or rule id")
	if err != nil {
		return nil, err
	}

	src := makeSource("policy or set or rule", ID, !ok, 0)

	rules, rOk := m[yastTagRules]
	policies, pOk := m[yastTagPolicies]
	effect, eOk := m[yastTagEffect]
	if rOk && pOk && eOk || rOk && pOk || rOk && eOk || pOk && eOk {
		tags := []string{}
		if rOk {
			tags = append(tags, yastTagRules)
		}

		if pOk {
			tags = append(tags, yastTagPolicies)
		}

		if eOk {
			tags = append(tags, yastTagEffect)
		}

		return nil, bindError(newEntityAmbiguityError(tags), src)
	}

	if rOk {
		return ctx.unmarshalPolicy(m, 0, ID, !ok, rules)
	}

	if pOk {
		return ctx.unmarshalPolicySet(m, 0, ID, !ok, policies)
	}

	if eOk {
		return ctx.unmarshalRuleEntity(m, ID, !ok, effect)
	}

	return nil, bindError(newEntityMissingKeyError(), src)
}

func (ctx *context) unmarshalCommand(v interface{}, u *pdp.PolicyUpdate) error {
	m, err := ctx.validateMap(v, "command")
	if err != nil {
		return err
	}

	s, err := ctx.extractString(m, yastTagOp, "operation")
	if err != nil {
		return err
	}

	op, ok := pdp.UpdateOpIDs[strings.ToLower(s)]
	if !ok {
		return newUnknownPolicyUpdateOperationError(s)
	}

	lst, err := ctx.extractList(m, yastTagPath, "path")
	if err != nil {
		return err
	}

	path := make([]string, len(lst))
	for i, item := range lst {
		s, ok := item.(string)
		if !ok {
			return newInvalidPolicyUpdatePathElementError(item, i+1)
		}

		path[i] = s
	}

	if op == pdp.UOAdd {
		entity, err := ctx.unmarshalEntity(m)
		if err != nil {
			return err
		}

		u.Append(op, path, entity)
	} else {
		u.Append(op, path, nil)
	}

	return nil
}
