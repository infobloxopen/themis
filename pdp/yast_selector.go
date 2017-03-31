package pdp

import "strings"

func (ctx *yastCtx) unmarshalSelectorPath(m map[interface{}]interface{}) (string, []interface{}, error) {
	path, err := ctx.extractString(m, yastTagPath, "selector path")
	if err != nil {
		return "", nil, err
	}

	p := make([]interface{}, 0)
	for i, item := range strings.Split(path, "/") {
		ID := strings.TrimPrefix(item, "$")
		if ID != item {
			a, ok := ctx.attrs[ID]
			if !ok {
				return path, nil, ctx.errorf("Unknown attribute ID %s for %d element of selector path", ID, i+1)
			}

			if a.DataType != DataTypeString && a.DataType != DataTypeDomain {
				return path, nil, ctx.errorf("Expected only %s or %s for %d element of selector path but got %s "+
					"attribute %s",
					DataTypeNames[DataTypeString], DataTypeNames[DataTypeDomain], i+1, DataTypeNames[a.DataType], ID)
			}

			p = append(p, AttributeDesignatorType{a})
			continue
		}

		p = append(p, item)
	}

	return path, p, nil
}

func (ctx *yastCtx) unmarshalSelectorContent(m map[interface{}]interface{}) (string, interface{}, error) {
	ID, err := ctx.extractString(m, yastTagContent, "selector content id")
	if err != nil {
		return "", nil, err
	}

	c, ok := ctx.includes[ID]
	if !ok {
		return ID, nil, ctx.errorf("No content with id %s", ID)
	}

	return ID, c, nil
}

func (ctx *yastCtx) unmarshalSelector(v interface{}) (*SelectorType, error) {
	ctx.pushNodeSpec(yastTagSelector)
	defer ctx.popNodeSpec()

	m, err := ctx.validateMap(v, "selector attributes")
	if err != nil {
		return nil, err
	}

	strT, err := ctx.extractString(m, yastTagType, "type")
	if err != nil {
		return nil, err
	}

	t, ok := DataTypeIDs[strings.ToLower(strT)]
	if !ok {
		return nil, ctx.errorf("Unknown value type %#v", strT)
	}

	strPath, rawPath, err := ctx.unmarshalSelectorPath(m)
	if err != nil {
		return nil, err
	}

	ID, rawCtx, err := ctx.unmarshalSelectorContent(m)
	if err != nil {
		return nil, err
	}

	if ctx.selectors == nil {
		ctx.selectors = make(map[string]map[string]*SelectorType)
	}

	pathMap, ok := ctx.selectors[ID]
	if !ok {
		pathMap = make(map[string]*SelectorType)
		ctx.selectors[ID] = pathMap
	}

	sel, ok := pathMap[strPath]
	if ok {
		return sel, nil
	}

	c, p, dp := prepareSelectorContent(rawCtx, rawPath, t)

	sel = &SelectorType{t, p, c, dp}
	pathMap[strPath] = sel

	return sel, nil
}
