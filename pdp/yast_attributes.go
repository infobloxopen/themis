package pdp

import "strings"

func (ctx *YastCtx) unmarshalAttributeDeclaration(k interface{}, v interface{}) error {
	ID, err := ctx.validateString(k, "attribute id")
	if err != nil {
		return err
	}

	ctx.pushNodeSpec("%#v", ID)
	defer ctx.popNodeSpec()

	strT, err := ctx.validateString(v, "attribute data type")
	if err != nil {
		return err
	}

	t, ok := DataTypeIDs[strings.ToLower(strT)]
	if !ok {
		return ctx.errorf("Expected attribute data type but got %#v", strT)
	}

	ctx.attrs[ID] = AttributeType{ID, t}
	return nil
}

func (ctx *YastCtx) unmarshalAttributeDeclarations(m map[interface{}]interface{}) error {
	ctx.attrs = make(map[string]AttributeType)

	attrs, err := ctx.extractMap(m, yastTagAttributes, "attribute declarations")
	if err != nil {
		return err
	}

	ctx.pushNodeSpec(yastTagAttributes)
	defer ctx.popNodeSpec()

	for k, v := range attrs {
		err := ctx.unmarshalAttributeDeclaration(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}
