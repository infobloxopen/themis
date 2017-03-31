package pdp

import (
	"fmt"
)

type yastCtx struct {
	nodeSpec  []string
	dataDir   string
	attrs     map[string]AttributeType
	includes  map[string]interface{}
	selectors map[string]map[string]*SelectorType
}

func newYASTCtx(dir string) yastCtx {
	return yastCtx{nodeSpec: []string{}, dataDir: dir}
}

func (ctx *yastCtx) pushNodeSpec(format string, a ...interface{}) {
	ctx.nodeSpec = append(ctx.nodeSpec, fmt.Sprintf(format, a...))
}

func (ctx *yastCtx) popNodeSpec() {
	if len(ctx.nodeSpec) < 1 {
		return
	}

	ctx.nodeSpec = ctx.nodeSpec[:len(ctx.nodeSpec)-1]
}

func (ctx yastCtx) validateString(v interface{}, desc string) (string, error) {
	r, ok := v.(string)
	if !ok {
		return "", ctx.errorf("Expected %s but got %T", desc, v)
	}

	return r, nil
}

func (ctx yastCtx) extractStringDef(m map[interface{}]interface{}, k string, def string, desc string) (string, error) {
	v, ok := m[k]
	if !ok {
		return def, nil
	}

	return ctx.validateString(v, desc)
}

func (ctx yastCtx) extractString(m map[interface{}]interface{}, k string, desc string) (string, error) {
	v, ok := m[k]
	if !ok {
		return "", ctx.errorf("Missing %s", desc)
	}

	return ctx.validateString(v, desc)
}

func (ctx yastCtx) validateMap(v interface{}, desc string) (map[interface{}]interface{}, error) {
	r, ok := v.(map[interface{}]interface{})
	if !ok {
		return nil, ctx.errorf("Expected %s but got %T", desc, v)
	}

	return r, nil
}

func (ctx yastCtx) extractMap(m map[interface{}]interface{}, k string, desc string) (map[interface{}]interface{}, error) {
	v, ok := m[k]
	if !ok {
		return nil, nil
	}

	return ctx.validateMap(v, desc)
}

func (ctx yastCtx) getSingleMapPair(m map[interface{}]interface{}, desc string) (interface{}, interface{}, error) {
	if len(m) > 1 {
		return nil, nil, ctx.errorf("Expected only one entry in %s got %d", desc, len(m))
	}

	for k, v := range m {
		return k, v, nil
	}

	return nil, nil, ctx.errorf("Expected at least one entry in %s got %d", desc, len(m))
}

func (ctx yastCtx) validateList(v interface{}, desc string) ([]interface{}, error) {
	r, ok := v.([]interface{})
	if !ok {
		return nil, ctx.errorf("Expected %s but got %T", desc, v)
	}

	return r, nil
}

func (ctx yastCtx) extractList(m map[interface{}]interface{}, k, desc string) ([]interface{}, error) {
	v, ok := m[k]
	if !ok {
		return nil, ctx.errorf("Missing %s", desc)
	}

	return ctx.validateList(v, desc)
}

func (ctx yastCtx) extractContentByItem(v interface{}) (interface{}, error) {
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

func (ctx yastCtx) extractStringOrMapDef(m map[interface{}]interface{}, k, defStr string, defMap map[interface{}]interface{}, desc string) (string, map[interface{}]interface{}, error) {
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
