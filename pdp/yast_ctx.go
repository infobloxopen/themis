package pdp

import (
	"fmt"
	"strings"
)

type ContentPolicyIndexItem struct {
	Path        []string
	SelectorMap map[interface{}]interface{}
}

type YastCtx struct {
	nodeSpec  []string
	dataDir   string
	attrs     map[string]AttributeType
	includes  map[string]interface{}
	selectors map[string]map[string]*SelectorType

	policyContentIdx   map[string]string
	contentPoliciesIdx map[string]map[string]ContentPolicyIndexItem
	policyIds          []string
}

func NewYASTCtx(dir string) YastCtx {
	return YastCtx{
		nodeSpec:           []string{},
		policyIds:          []string{},
		policyContentIdx:   map[string]string{},
		contentPoliciesIdx: map[string]map[string]ContentPolicyIndexItem{},
		dataDir:            dir,
	}
}

func (ctx *YastCtx) Reset() {
	ctx.nodeSpec = []string{}
	ctx.includes = nil
	ctx.selectors = nil
	ctx.policyIds = []string{}
}

func (ctx *YastCtx) SetContent(c map[string]interface{}) {
	ctx.includes = c
}

func (ctx *YastCtx) RemovePolicyFromContentIndex() {
	if len(ctx.policyIds) == 0 {
		return
	}

	// Remove policy(set) and all sub-policies(sets) from the index.
	pkey := ctx.PolicyIndexKey(ctx.policyIds)
	for pk, ck := range ctx.policyContentIdx {
		if strings.HasPrefix(pk, pkey) {
			if pmap, ok := ctx.contentPoliciesIdx[ck]; ok {
				delete(pmap, pk)
				if len(pmap) == 0 {
					delete(ctx.contentPoliciesIdx, ck)
				}
			}
			delete(ctx.policyContentIdx, pk)
		}
	}
}

func (ctx *YastCtx) addPolicyToContentIndex(ckey string, smap map[interface{}]interface{}) {
	pkey := ctx.PolicyIndexKey(ctx.policyIds)
	pids := make([]string, len(ctx.policyIds), len(ctx.policyIds))
	copy(pids, ctx.policyIds)
	ctx.policyContentIdx[pkey] = ckey
	ii := ContentPolicyIndexItem{Path: pids, SelectorMap: smap}
	pmap, ok := ctx.contentPoliciesIdx[ckey]
	if ok {
		pmap[pkey] = ii
	} else {
		pmap = map[string]ContentPolicyIndexItem{pkey: ii}
	}
	ctx.contentPoliciesIdx[ckey] = pmap
}

func (ctx *YastCtx) PolicyIndexKey(path []string) string {
	return strings.Join(path, "/")
}

func (ctx *YastCtx) PoliciesFromContentIndex(cpath []string) map[string]ContentPolicyIndexItem {
	parts := make([]string, len(cpath))
	for i, v := range cpath {
		parts[i] = fmt.Sprintf("%q", v)
	}
	ckey := strings.Join(parts, "/")
	if pmap, ok := ctx.contentPoliciesIdx[ckey]; ok {
		return pmap
	}
	return map[string]ContentPolicyIndexItem{}
}

func (ctx *YastCtx) PushPolicyID(id string) {
	ctx.policyIds = append(ctx.policyIds, id)
}

func (ctx *YastCtx) PopPolicyID() {
	if len(ctx.policyIds) < 1 {
		return
	}

	ctx.policyIds = ctx.policyIds[:len(ctx.policyIds)-1]
}

func (ctx *YastCtx) UpdateEvaluableTypeContent(e EvaluableType, meta interface{}) error {
	switch p := e.(type) {
	case *PolicySetType:
		if mpca, ok := p.AlgParams.(*MapperPCAParams); ok {
			if s, ok := mpca.Argument.(*SelectorType); !ok {
				return ctx.errorf("Expected %T but got %T", s, mpca.Argument)
			}

			s, err := ctx.unmarshalSelector(meta)
			if err != nil {
				return err
			}

			mpca.Argument = s
		} else {
			return ctx.errorf("Expected %T but got %T", mpca, p.AlgParams)
		}
	case *PolicyType:
		if mrca, ok := p.AlgParams.(*MapperRCAParams); ok {
			if s, ok := mrca.Argument.(*SelectorType); !ok {
				return ctx.errorf("Expected %T but got %T", s, mrca.Argument)
			}

			s, err := ctx.unmarshalSelector(meta)
			if err != nil {
				return err
			}

			mrca.Argument = s
		} else {
			return ctx.errorf("Expected %T but got %T", mrca, p.AlgParams)
		}
	default:
		return ctx.errorf("Unsupported evaluable type %T", p)
	}

	return nil
}

func (ctx *YastCtx) pushNodeSpec(format string, a ...interface{}) {
	ctx.nodeSpec = append(ctx.nodeSpec, fmt.Sprintf(format, a...))
}

func (ctx *YastCtx) popNodeSpec() {
	if len(ctx.nodeSpec) < 1 {
		return
	}

	ctx.nodeSpec = ctx.nodeSpec[:len(ctx.nodeSpec)-1]
}

func (ctx YastCtx) validateString(v interface{}, desc string) (string, error) {
	r, ok := v.(string)
	if !ok {
		return "", ctx.errorf("Expected %s but got %T", desc, v)
	}

	return r, nil
}

func (ctx YastCtx) extractStringDef(m map[interface{}]interface{}, k string, def string, desc string) (string, error) {
	v, ok := m[k]
	if !ok {
		return def, nil
	}

	return ctx.validateString(v, desc)
}

func (ctx YastCtx) extractString(m map[interface{}]interface{}, k string, desc string) (string, error) {
	v, ok := m[k]
	if !ok {
		return "", ctx.errorf("Missing %s", desc)
	}

	return ctx.validateString(v, desc)
}

func (ctx YastCtx) validateMap(v interface{}, desc string) (map[interface{}]interface{}, error) {
	r, ok := v.(map[interface{}]interface{})
	if !ok {
		return nil, ctx.errorf("Expected %s but got %T", desc, v)
	}

	return r, nil
}

func (ctx YastCtx) extractMap(m map[interface{}]interface{}, k string, desc string) (map[interface{}]interface{}, error) {
	v, ok := m[k]
	if !ok {
		return nil, nil
	}

	return ctx.validateMap(v, desc)
}

func (ctx YastCtx) getSingleMapPair(m map[interface{}]interface{}, desc string) (interface{}, interface{}, error) {
	if len(m) > 1 {
		return nil, nil, ctx.errorf("Expected only one entry in %s got %d", desc, len(m))
	}

	for k, v := range m {
		return k, v, nil
	}

	return nil, nil, ctx.errorf("Expected at least one entry in %s got %d", desc, len(m))
}

func (ctx YastCtx) validateList(v interface{}, desc string) ([]interface{}, error) {
	r, ok := v.([]interface{})
	if !ok {
		return nil, ctx.errorf("Expected %s but got %T", desc, v)
	}

	return r, nil
}

func (ctx YastCtx) extractList(m map[interface{}]interface{}, k, desc string) ([]interface{}, error) {
	v, ok := m[k]
	if !ok {
		return nil, ctx.errorf("Missing %s", desc)
	}

	return ctx.validateList(v, desc)
}

func (ctx YastCtx) extractContentByItem(v interface{}) (interface{}, error) {
	ID, err := ctx.validateString(v, "")
	if err != nil {
		return nil, nil
	}

	c, ok := ctx.includes[ID]
	if !ok {
		return nil, ctx.errorf("No content with id %s", ID)
	}

	return c, nil
}

func (ctx YastCtx) extractStringOrMapDef(m map[interface{}]interface{}, k, defStr string, defMap map[interface{}]interface{}, desc string) (string, map[interface{}]interface{}, error) {
	v, ok := m[k]
	if !ok {
		return defStr, defMap, nil
	}

	switch r := v.(type) {
	case string:
		return r, nil, nil
	case map[interface{}]interface{}:
		return "", r, nil
	}

	return "", nil, ctx.errorf("Expected %s but got %T", desc, v)
}
