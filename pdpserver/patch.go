package main

import (
	"fmt"

	"github.com/infobloxopen/themis/pdp"

	"gopkg.in/yaml.v2"
)

const (
	PatchOpAppend = "APPEND"
	PatchOpDelete = "DELETE"
)

type PatchItem struct {
	Op     string
	Path   []string
	Entity interface{}

	pathIndex int
}

func (pi *PatchItem) getID() string {
	return pi.Path[pi.pathIndex]
}

func (pi *PatchItem) nextID() bool {
	if len(pi.Path) == pi.pathIndex+1 {
		return false
	}

	pi.pathIndex++
	return true
}

type patchCtx struct {
	curOriginalNode interface{}
	curPatchedNode  interface{}
	prevPatchedNode interface{}

	pi *PatchItem
}

func copyItem(item interface{}) interface{} {
	switch item := item.(type) {
	case pdp.PolicySetType, pdp.PolicyType, pdp.RuleType:
		return item
	default:
		panic(fmt.Sprintf("Unexpected item type %T", item))
	}
}

func findItem(id string, parent interface{}) (bool, int, interface{}) {
	switch parent := parent.(type) {
	case pdp.PolicySetType:
		for i, e := range parent.Policies {
			if e.GetID() == id {
				return true, i, e
			}
		}
	case pdp.PolicyType:
		for i, r := range parent.Rules {
			if r.ID == id {
				return true, i, r
			}
		}
	default:
		panic(fmt.Sprintf("Unexpected node type %T", parent))
	}

	return false, -1, nil
}

func (s *Server) applyPatchItem(ctx *patchCtx) error {
	pi := ctx.pi
	id := pi.getID()

	if pi.nextID() {
		var (
			onext, pnext interface{}
			ok           bool
		)

		nid := pi.getID()

		if ok, _, onext = findItem(nid, ctx.curOriginalNode); !ok {
			return fmt.Errorf("Cannot find %s next node in original %s node", nid, id)
		}

		if ok, _, pnext = findItem(nid, ctx.curPatchedNode); !ok {
			return fmt.Errorf("Cannot find %s next node in patched %s node", nid, id)
		}

		if &onext == &pnext {
			pnext = copyItem(onext)
		}

		ctx.curOriginalNode = onext
		ctx.prevPatchedNode = ctx.curPatchedNode
		ctx.curPatchedNode = pnext

		return s.applyPatchItem(ctx)
	}

	op := pi.Op

	switch op {
	case PatchOpAppend:
		switch p := ctx.curPatchedNode.(type) {
		case pdp.PolicySetType:
			e, err := s.ctx.UnmarshalEvaluable(pi.Entity)
			if err != nil {
				return err
			}

			if ok, _, _ := findItem(e.GetID(), p); ok {
				return fmt.Errorf("Cannot append %s node to patched %s node. Node already exists", e.GetID(), id)
			}

			p.Policies = append(p.Policies, e)
		case pdp.PolicyType:
			rule, err := s.ctx.UnmarshalRule(pi.Entity)
			if err != nil {
				return err
			}

			if ok, _, _ := findItem(rule.ID, p); ok {
				return fmt.Errorf("Cannot append %s rule to patched %s policy. Rule already exists", rule.ID, id)
			}

			p.Rules = append(p.Rules, rule)
		default:
			return fmt.Errorf("Operation %s is unsupported for type %T", op, p)
		}
	case PatchOpDelete:
		switch p := ctx.curPatchedNode.(type) {
		case pdp.PolicySetType:
			prev := ctx.prevPatchedNode.(pdp.PolicySetType)

			ok, i, _ := findItem(id, prev)
			if !ok {
				return fmt.Errorf("Cannot delete %s policy set from patched %s policy set. Policy set does not exist", id, prev.ID)
			}

			prev.Policies = append(prev.Policies[:i], prev.Policies[i+1:]...)
		case pdp.PolicyType:
			prev := ctx.prevPatchedNode.(pdp.PolicySetType)

			ok, i, _ := findItem(id, prev)
			if !ok {
				return fmt.Errorf("Cannot delete %s policy from patched %s policy set. Policy does not exist", id, prev.ID)
			}

			prev.Policies = append(prev.Policies[:i], prev.Policies[i+1:]...)
		case pdp.RuleType:
			prev := ctx.prevPatchedNode.(pdp.PolicyType)

			ok, i, _ := findItem(id, prev)
			if !ok {
				return fmt.Errorf("Cannot delete %s rule from patched %s policy. Rule does not exist", id, prev.ID)
			}

			prev.Rules = append(prev.Rules[:i], prev.Rules[i+1:]...)
		default:
			return fmt.Errorf("Operation %s is unsupported for type %T", op, p)
		}
	default:
		return fmt.Errorf("Unsupported %s patch operation", op)
	}

	return nil
}

func (s *Server) makePatchedPolicies(data []byte, content map[string]interface{}) (pdp.EvaluableType, error) {
	var patches []PatchItem
	if err := yaml.Unmarshal(data, &patches); err != nil {
		return nil, err
	}

	patchedPolicy := copyItem(s.Policy)
	for _, pi := range patches {
		pctx := &patchCtx{
			curOriginalNode: s.Policy,
			curPatchedNode:  patchedPolicy,
			prevPatchedNode: nil,
			pi:              &pi,
		}

		if err := s.applyPatchItem(pctx); err != nil {
			return nil, err
		}
	}

	return patchedPolicy.(pdp.EvaluableType), nil
}

func (s *Server) makePatchedContent(data []byte, id string) (interface{}, error) {
	return s.Includes, nil
}
