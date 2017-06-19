package pdp

import "testing"

func TestUnmarshalExpressionValueByType(t *testing.T) {
	ctx := newYASTCtx("")
	ctx.includes = map[string]interface{}{
		"ListOfStrings": []interface{}{"test", "example", "end"}}

	v, err := ctx.unmarshalValueByType(DataTypeUndefined, nil)
	if err == nil {
		t.Errorf("Expected error for unmarshalling of %s but got result %T (%#v)",
			DataTypeNames[DataTypeUndefined], v, v)
	}

	v, err = ctx.unmarshalValueByType(DataTypeListOfStrings,
		[]interface{}{"test", "example", "end"})
	if err != nil {
		t.Errorf("Expected result for unmarshaling of %s but got error: %s",
			DataTypeNames[DataTypeListOfStrings], err)
	}

	if v.DataType != DataTypeListOfStrings {
		t.Errorf("Expected attribute value %s but got %s",
			DataTypeNames[DataTypeListOfStrings], DataTypeNames[v.DataType])
	}

	desc := v.describe()
	if desc != "[test, example, end]" {
		t.Errorf("Expected attribute value description %s but got %s", "[test, example, end]", desc)
	}

	v, err = ctx.unmarshalValueByType(DataTypeListOfStrings, "ListOfStrings")
	if err != nil {
		t.Errorf("Expected result for unmarshaling of %s from content but got error: %s",
			DataTypeNames[DataTypeListOfStrings], err)
	}

	if v.DataType != DataTypeListOfStrings {
		t.Errorf("Expected attribute value %s from content but got %s",
			DataTypeNames[DataTypeListOfStrings], DataTypeNames[v.DataType])
	}

	desc = v.describe()
	if desc != "[test, example, end]" {
		t.Errorf("Expected attribute value description %s from content but got %s", "[test, example, end]", desc)
	}
}
