package main

import (
	"encoding/json"
	"fmt"

	log "github.com/Sirupsen/logrus"
)

type contentPatchCtx struct {
	// Current item in the patched tree.
	cur interface{}
	// Parent (previous) item in the patched tree.
	prev interface{}
	// Content ID.
	cid string

	pi *PatchItem
}

func (s *Server) trackAffectedPolicies(ctx *contentPatchCtx) {
	contentPath := []string{ctx.cid}
	if len(ctx.pi.Path) > 0 {
		contentPath = append(contentPath, ctx.pi.Path[:ctx.pi.pathIndex]...)
	}
	pmap := s.ctx.PoliciesFromContentIndex(contentPath)
	for k, v := range pmap {
		s.AffectedPolicies[k] = v
	}
}

func (s *Server) applyContentPatchItem(ctx *contentPatchCtx) error {
	pi := ctx.pi
	id := pi.getID()

	s.trackAffectedPolicies(ctx)

	if pi.nextID() {
		var (
			next interface{}
			ok   bool
		)

		curmap, ok := ctx.cur.(map[string]interface{})
		if !ok {
			return fmt.Errorf("Unsupported content item type '%T'", curmap)
		}

		next, ok = curmap[id]
		if !ok {
			return fmt.Errorf("Cannot find '%v' next item in '%s' content", pi.Path[:pi.pathIndex], ctx.cid)
		}

		ctx.prev = ctx.cur
		ctx.cur = next

		return s.applyContentPatchItem(ctx)
	}

	switch pi.Op {
	case PatchOpAdd:
		switch cur := ctx.cur.(type) {
		case map[string]interface{}:
			cur[id] = pi.Entity
		case []string:
			e, ok := pi.Entity.(string)
			if !ok {
				return fmt.Errorf("Cannot add '%T' item to list of strings of '%s' content", pi.Entity, ctx.cid)
			}

			cur = append(cur, e)
			if ctx.prev != nil {
				pmap := ctx.prev.(map[string]interface{})
				pid := pi.Path[len(pi.Path)-1]
				pmap[pid] = cur
			}
		default:
			return fmt.Errorf("Operation '%s' is unsupported for type '%T'", pi.Op, cur)
		}
	case PatchOpDelete:
		switch cur := ctx.cur.(type) {
		case map[string]interface{}:
			if _, ok := cur[id]; !ok {
				return fmt.Errorf("Cannot delete '%s' item from '%s' content. Item does not exist", pi.Path, ctx.cid)
			}

			delete(cur, id)
		case []string:
			var (
				i int
				v string
			)

			for i, v = range cur {
				if v == id {
					break
				}
			}

			if i == len(cur) {
				return fmt.Errorf("Cannot delete '%s' item from '%s' content. Item does not exist", pi.Path, ctx.cid)
			}

			cur = append(cur[:i], cur[i+1:]...)
			if ctx.prev != nil {
				pmap := ctx.prev.(map[string]interface{})
				pid := pi.Path[len(pi.Path)-2]
				pmap[pid] = cur
			}
		}
	default:
		return fmt.Errorf("Unsupported '%s' patch operation for content", pi.Op)
	}

	return nil
}

func (s *Server) patchContent(data []byte, id string) (interface{}, error) {
	var patches []PatchItem
	if err := json.Unmarshal(data, &patches); err != nil {
		return nil, err
	}

	content := s.Includes[id]
	for _, pi := range patches {
		log.Debugf("Applying patch operation to '%s' content: %+v", id, pi)

		ctx := &contentPatchCtx{
			cid:  id,
			cur:  content,
			prev: nil,
			pi:   &pi,
		}

		if err := s.applyContentPatchItem(ctx); err != nil {
			return nil, err
		}
	}

	return content, nil
}
