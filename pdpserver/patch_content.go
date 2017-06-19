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

func (ctx *contentPatchCtx) errorf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("Failed to apply [%s:%v] for content '%s': %s", ctx.pi.Op, ctx.pi.Path, ctx.cid, msg)
}

func (s *Server) trackAffectedPolicies(ctx *contentPatchCtx) {
	contentPath := []string{ctx.cid}
	if len(ctx.pi.Path) > 0 {
		contentPath = append(contentPath, ctx.pi.Path[:ctx.pi.pathIndex]...)
	}
	pmap := s.Ctx.PoliciesFromContentIndex(contentPath)
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
			return ctx.errorf("Unsupported content item type '%T'", curmap)
		}

		next, ok = curmap[id]
		if !ok {
			return ctx.errorf("Cannot find '%v' next item", pi.Path[:pi.pathIndex])
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
		case []interface{}:
			e, ok := pi.Entity.(string)
			if !ok {
				return ctx.errorf("Cannot add '%T' item to list of strings", pi.Entity)
			}

			cur = append(cur, e)
			if ctx.prev != nil {
				pmap := ctx.prev.(map[string]interface{})
				pid := pi.Path[len(pi.Path)-1]
				pmap[pid] = cur
			}
		default:
			return ctx.errorf("Operation is unsupported for type '%T'", cur)
		}
	case PatchOpDelete:
		switch cur := ctx.cur.(type) {
		case map[string]interface{}:
			if _, ok := cur[id]; !ok {
				return ctx.errorf("Cannot delete item. Item does not exist")
			}

			delete(cur, id)
		case []interface{}:
			var (
				i int
				v interface{}
			)

			for i, v = range cur {
				if v == id {
					break
				}
			}

			if i == len(cur) {
				return ctx.errorf("Cannot delete item. Item does not exist")
			}

			cur = append(cur[:i], cur[i+1:]...)
			if ctx.prev != nil {
				pmap := ctx.prev.(map[string]interface{})
				pid := pi.Path[len(pi.Path)-2]
				pmap[pid] = cur
			}
		}
	default:
		return ctx.errorf("Unsupported patch operation for content")
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
