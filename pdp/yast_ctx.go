package pdp

import (
	"fmt"
)

type YastCtx struct {
	nodeSpec  []string
	dataDir   string
	attrs     map[string]AttributeType
	includes  map[string]interface{}
	selectors map[string]map[string]*SelectorType
}

func NewYASTCtx(dir string) YastCtx {
	return YastCtx{nodeSpec: []string{}, dataDir: dir}
}

func (ctx *YastCtx) Reset() {
	ctx.nodeSpec = []string{}
	ctx.includes = nil
	ctx.selectors = nil
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
