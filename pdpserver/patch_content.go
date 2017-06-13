package main

import (
	"encoding/json"
	"fmt"
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

func unmarshalContentEntity(e interface{}) interface{} {
	switch e := e.(type) {
	case map[interface{}]interface{}, []string, string:
		return e
	case map[string]interface{}:
		cpy := make(map[interface{}]interface{}, len(e))
		for k, v := range e {
			cpy[k] = v
		}
		return cpy
	case []interface{}:
		cpy := make([]string, len(e), len(e))
		for i, v := range e {
			cpy[i] = v.(string)
		}
		return cpy
	default:
		panic(fmt.Sprintf("Unsupported content entity type '%T'", e))
	}
}

func (s *Server) applyContentPatchItem(ctx *contentPatchCtx) error {
	pi := ctx.pi
	id := pi.getID()

	if pi.nextID() {
		var (
			next interface{}
			ok   bool
		)

		switch curmap := ctx.cur.(type) {
		case map[interface{}]map[interface{}]interface{}:
			if next, ok = curmap[id]; !ok {
				return fmt.Errorf("Cannot find '%v' next item in '%s' content", pi.Path[:pi.pathIndex], ctx.cid)
			}

			ctx.prev = ctx.cur
			ctx.cur = next

		case map[interface{}]interface{}:
			if next, ok = curmap[id]; !ok {
				return fmt.Errorf("Cannot find '%v' next item in '%' content", pi.Path[:pi.pathIndex], ctx.cid)
			}

			ctx.prev = ctx.cur
			ctx.cur = next
		default:
			return fmt.Errorf("Unsupported content item type '%T'", curmap)
		}

		return s.applyContentPatchItem(ctx)
	}

	switch pi.Op {
	case PatchOpAdd:
		switch item := ctx.cur.(type) {
		case map[interface{}]map[interface{}]interface{}:
			entity, ok := unmarshalContentEntity(pi.Entity).(map[interface{}]interface{})
			if !ok {
				return fmt.Errorf("Cannot add '%T' item to '%T' item for '%s' content.", pi.Entity, item, ctx.cid)
			}

			item[id] = entity
		case map[interface{}]interface{}:
			item[id] = unmarshalContentEntity(pi.Entity)
		default:
			return fmt.Errorf("Operation '%s' is unsupported for type '%T'", pi.Op, item)
		}
	case PatchOpDelete:
		switch item := ctx.cur.(type) {
		case map[interface{}]map[interface{}]interface{}:
			if _, ok := item[id]; !ok {
				return fmt.Errorf("Cannot delete '%s' item from '%s' content. Item does not exist", pi.Path, ctx.cid)
			}
			delete(item, id)
		case map[interface{}]interface{}:
			if _, ok := item[id]; !ok {
				return fmt.Errorf("Cannot delete '%s' item from '%s' content. Item does not exist", pi.Path, ctx.cid)
			}
			delete(item, id)
		case []string:
			var (
				i int
				v string
			)

			for i, v = range item {
				if v == id {
					break
				}
			}

			if i == len(item) {
				return fmt.Errorf("Cannot delete '%s' item from '%s' content. Item does not exist", pi.Path, ctx.cid)
			}

			item = append(item[:i], item[i+1:]...)

			pmap := ctx.prev.(map[interface{}]interface{})
			pid := pi.Path[len(pi.Path)-2]
			pmap[pid] = item
		case string:
			pmap := ctx.prev.(map[interface{}]interface{})
			if _, ok := pmap[id]; !ok {
				return fmt.Errorf("Cannot delete '%s' item from '%s' content. Item does not exist", pi.Path, ctx.cid)
			}

			delete(pmap, id)
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
