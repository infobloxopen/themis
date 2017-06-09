package pdp

func (ctx *YastCtx) unmarshalObligationItem(v interface{}, i int, exprs []AttributeAssignmentExpressionType) ([]AttributeAssignmentExpressionType, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	o, err := ctx.validateMap(v, "obligations")
	if err != nil {
		return nil, err
	}

	k, v, err := ctx.getSingleMapPair(o, "obligation map")
	if err != nil {
		return nil, err
	}

	ID, err := ctx.validateString(k, "obligation attribute id")
	if err != nil {
		return nil, err
	}

	a, ok := ctx.attrs[ID]
	if !ok {
		return nil, ctx.errorf("Unknown attribute ID %s", ID)
	}

	ctx.pushNodeSpec("%#v", ID)
	defer ctx.popNodeSpec()

	e, err := ctx.unmarshalValueByType(a.DataType, v)
	if err != nil {
		return nil, err
	}

	return append(exprs, AttributeAssignmentExpressionType{AttributeType{ID, e.DataType}, *e}), nil
}

func (ctx *YastCtx) unmarshalObligation(m map[interface{}]interface{}) ([]AttributeAssignmentExpressionType, error) {
	root, ok := m[yastTagObligation]
	if !ok {
		return []AttributeAssignmentExpressionType{}, nil
	}

	ctx.pushNodeSpec(yastTagObligation)
	defer ctx.popNodeSpec()

	items, err := ctx.validateList(root, "obligations")
	if err != nil {
		return []AttributeAssignmentExpressionType{}, err
	}

	var r []AttributeAssignmentExpressionType

	for i, item := range items {
		r, err = ctx.unmarshalObligationItem(item, i, r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}
