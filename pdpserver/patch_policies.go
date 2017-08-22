package main

import (
	"fmt"

	"github.com/infobloxopen/themis/pdp"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	PatchOpAdd     = "add"
	PatchOpDelete  = "delete"
	PatchOpRefresh = "refresh"
)

type PatchItem struct {
	Op     string
	Path   []string
	Entity interface{}

	pathIndex int
}

func (pi *PatchItem) getID() string {
	if len(pi.Path) == 0 {
		return ""
	}
	return pi.Path[pi.pathIndex]
}

func (pi *PatchItem) nextID() bool {
	if len(pi.Path) <= pi.pathIndex+1 {
		return false
	}

	pi.pathIndex++
	return true
}

type policiesPatchCtx struct {
	// Current item in the original tree.
	ocur interface{}
	// Current item in the patched tree.
	cur interface{}
	// Parent (previous) item in the patched tree.
	prev interface{}

	pi *PatchItem
}

func (ctx *policiesPatchCtx) errorf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("Failed to apply [%s:%v]: %s", ctx.pi.Op, ctx.pi.Path, msg)
}

func (s *Server) untrackAffectedPolicies(path []string) {
	k := s.Ctx.PolicyIndexKey(path)
	delete(s.AffectedPolicies, k)
}

func copyAndUpdateAlg(e pdp.EvaluableType) {
	switch e := e.(type) {
	case *pdp.PolicySetType:
		if e.AlgParams != nil {
			m := e.AlgParams.(*pdp.MapperPCAParams)
			mcpy := *m
			pmap := make(map[string]pdp.EvaluableType)
			for _, p := range e.Policies {
				pmap[p.GetID()] = p
				if mcpy.DefaultPolicy != nil && mcpy.DefaultPolicy.GetID() == p.GetID() {
					mcpy.DefaultPolicy = p
				}
				if mcpy.ErrorPolicy != nil && mcpy.ErrorPolicy.GetID() == p.GetID() {
					mcpy.ErrorPolicy = p
				}
			}
			mcpy.PoliciesMap = pmap
			e.AlgParams = &mcpy
		}
	case *pdp.PolicyType:
		if e.AlgParams != nil {
			m := e.AlgParams.(*pdp.MapperRCAParams)
			mcpy := *m
			rmap := make(map[string]*pdp.RuleType)
			for _, r := range e.Rules {
				rmap[r.ID] = r
				if mcpy.DefaultRule != nil && mcpy.DefaultRule.ID == r.ID {
					mcpy.DefaultRule = r
				}
				if mcpy.ErrorRule != nil && mcpy.ErrorRule.ID == r.ID {
					mcpy.ErrorRule = r
				}
			}
			mcpy.RulesMap = rmap
			e.AlgParams = &mcpy
		}
	default:
		panic(fmt.Sprintf("Unsupported policies item type '%T'", e))
	}
}

func copyPoliciesItem(item interface{}, parent interface{}, i int) interface{} {
	switch item := item.(type) {
	case *pdp.PolicySetType:
		cpy := *item
		cpy.Policies = make([]pdp.EvaluableType, len(item.Policies), len(item.Policies))
		copy(cpy.Policies, item.Policies)
		if parent != nil {
			ppset := parent.(*pdp.PolicySetType)
			ppset.Policies[i] = &cpy
			copyAndUpdateAlg(ppset)
		}
		return &cpy
	case *pdp.PolicyType:
		cpy := *item
		cpy.Rules = make([]*pdp.RuleType, len(item.Rules), len(item.Rules))
		copy(cpy.Rules, item.Rules)
		if parent != nil {
			ppset := parent.(*pdp.PolicySetType)
			ppset.Policies[i] = &cpy
			copyAndUpdateAlg(ppset)
		}
		return &cpy
	case *pdp.RuleType:
		cpy := *item
		if parent != nil {
			pp := parent.(*pdp.PolicyType)
			pp.Rules[i] = &cpy
			copyAndUpdateAlg(pp)
		}
		return &cpy
	default:
		panic(fmt.Sprintf("Unsupported policies item type '%T'", item))
	}
}

func findPoliciesItem(id string, parent interface{}) (bool, int, interface{}) {
	switch parent := parent.(type) {
	case *pdp.PolicySetType:
		for i, e := range parent.Policies {
			if e.GetID() == id {
				return true, i, e
			}
		}
	case *pdp.PolicyType:
		for i, r := range parent.Rules {
			if r.ID == id {
				return true, i, r
			}
		}
	default:
		panic(fmt.Sprintf("Unsupported policies item type '%T'", parent))
	}

	return false, -1, nil
}

func (s *Server) applyPoliciesPatchItem(ctx *policiesPatchCtx) error {
	pi := ctx.pi
	id := pi.getID()

	if _, ok := ctx.cur.(pdp.EvaluableType); ok {
		s.Ctx.PushPolicyID(id)
		defer s.Ctx.PopPolicyID()
	}

	if pi.nextID() {
		var (
			onext, next interface{}
			ok          bool
			i           int
		)

		nid := pi.getID()

		// Patch into patch is allowed.
		if ctx.ocur != nil {
			_, _, onext = findPoliciesItem(nid, ctx.ocur)
		}

		if ok, i, next = findPoliciesItem(nid, ctx.cur); !ok {
			return ctx.errorf("Cannot find '%v' next item in patched content", pi.Path[:pi.pathIndex+1])
		}

		if onext == next {
			next = copyPoliciesItem(next, ctx.cur, i)
		}

		ctx.ocur = onext
		ctx.prev = ctx.cur
		ctx.cur = next

		return s.applyPoliciesPatchItem(ctx)
	}

	switch pi.Op {
	case PatchOpAdd:
		switch item := ctx.cur.(type) {
		case *pdp.PolicySetType:
			id, err := s.Ctx.UnmarshalEvaluableID(pi.Entity)
			if err != nil {
				return ctx.errorf("%v", err)
			}

			s.Ctx.PushPolicyID(id)
			s.Ctx.RemovePolicyFromContentIndex()
			s.Ctx.PopPolicyID()

			e, err := s.Ctx.UnmarshalEvaluable(pi.Entity)
			if err != nil {
				return ctx.errorf("%v", err)
			}

			s.untrackAffectedPolicies(append(pi.Path, e.GetID()))

			item.Policies = append(item.Policies, e)
			copyAndUpdateAlg(item)
		case *pdp.PolicyType:
			r, err := s.Ctx.UnmarshalRule(pi.Entity)
			if err != nil {
				return ctx.errorf("%v", err)
			}

			item.Rules = append(item.Rules, r)
			copyAndUpdateAlg(item)
		default:
			return ctx.errorf("Operation is unsupported for type '%T'", item)
		}
	case PatchOpDelete:
		switch item := ctx.cur.(type) {
		case *pdp.PolicySetType:
			ppset := ctx.prev.(*pdp.PolicySetType)

			ok, i, _ := findPoliciesItem(id, ppset)
			if !ok {
				return ctx.errorf("Cannot delete policy set from patched policy set. Policy set does not exist")
			}

			s.Ctx.RemovePolicyFromContentIndex()
			s.untrackAffectedPolicies(pi.Path)

			ppset.Policies = append(ppset.Policies[:i], ppset.Policies[i+1:]...)
			copyAndUpdateAlg(ppset)
		case *pdp.PolicyType:
			ppset := ctx.prev.(*pdp.PolicySetType)

			ok, i, _ := findPoliciesItem(id, ppset)
			if !ok {
				return ctx.errorf("Cannot delete policy from patched policy set. Policy does not exist")
			}

			s.Ctx.RemovePolicyFromContentIndex()
			s.untrackAffectedPolicies(pi.Path)

			ppset.Policies = append(ppset.Policies[:i], ppset.Policies[i+1:]...)
			copyAndUpdateAlg(ppset)
		case *pdp.RuleType:
			pp := ctx.prev.(*pdp.PolicyType)

			ok, i, _ := findPoliciesItem(id, pp)
			if !ok {
				return ctx.errorf("Cannot delete rule from patched policy. Rule does not exist")
			}

			pp.Rules = append(pp.Rules[:i], pp.Rules[i+1:]...)
			copyAndUpdateAlg(pp)
		default:
			return ctx.errorf("Operation is unsupported for type '%T'", item)
		}
	case PatchOpRefresh:
		switch item := ctx.cur.(type) {
		case pdp.EvaluableType:
			copyAndUpdateAlg(item)
			if err := s.Ctx.UpdateEvaluableTypeContent(item, pi.Entity); err != nil {
				return err
			}
		default:
			return ctx.errorf("Operation is unsupported for type '%T'", item)
		}
	default:
		return ctx.errorf("Unsupported patch operation for policies")
	}

	return nil
}

func (s *Server) makePoliciesRefreshPatches() []PatchItem {
	patches := make([]PatchItem, 0, len(s.AffectedPolicies))
	for _, v := range s.AffectedPolicies {
		pi := PatchItem{Op: PatchOpRefresh, Path: v.Path, Entity: v.SelectorMap}
		patches = append(patches, pi)
	}
	return patches
}

func (s *Server) copyAndPatchPolicies(data []byte, content map[string]interface{}) (pdp.EvaluableType, error) {
	var patches []PatchItem
	if err := yaml.Unmarshal(data, &patches); err != nil {
		return nil, err
	}

	s.Ctx.SetContent(content)
	cpyPolicy := copyPoliciesItem(s.Policy, nil, 0)
	for _, pi := range patches {
		log.Debugf("Applying patch operation to policies: %+v", pi)

		ctx := &policiesPatchCtx{
			ocur: s.Policy,
			cur:  cpyPolicy,
			prev: nil,
			pi:   &pi,
		}

		if err := s.applyPoliciesPatchItem(ctx); err != nil {
			return nil, err
		}
	}

	for _, pi := range s.makePoliciesRefreshPatches() {
		log.Debugf("Update policies content: %+v", pi)

		ctx := &policiesPatchCtx{
			ocur: s.Policy,
			cur:  cpyPolicy,
			prev: nil,
			pi:   &pi,
		}

		if err := s.applyPoliciesPatchItem(ctx); err != nil {
			return nil, err
		}
	}

	return cpyPolicy.(pdp.EvaluableType), nil
}
