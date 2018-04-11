package pdp

import "testing"

func TestAttributeDesignator(t *testing.T) {
	testAttributes := []struct {
		id  string
		val AttributeValue
	}{
		{
			id:  "test-id",
			val: MakeStringValue("test-value"),
		},
		{
			id:  "test-id-i",
			val: MakeIntegerValue(12345),
		},
		{
			id:  "test-id-f",
			val: MakeFloatValue(67.89),
		},
	}

	ctx, err := NewContext(nil, len(testAttributes), func(i int) (string, AttributeValue, error) {
		return testAttributes[i].id, testAttributes[i].val, nil
	})
	if err != nil {
		t.Fatalf("Expected context but got error %s", err)
	}

	a := Attribute{
		id: "test-id",
		t:  TypeString}
	d := MakeAttributeDesignator(a)
	dai := d.GetID()
	if dai != "test-id" {
		t.Errorf("Expected %q id but got %q", "test-id", dai)
	}

	dat := d.GetResultType()
	if dat != TypeString {
		t.Errorf("Expected %q type but got %q", BuiltinTypeNames[TypeString], BuiltinTypeNames[dat])
	}

	_, err = d.Calculate(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	i := Attribute{
		id: "test-id-i",
		t:  TypeInteger}
	d = MakeAttributeDesignator(i)
	dat = d.GetResultType()
	if dat != TypeInteger {
		t.Errorf("Expected %q type but got %q", BuiltinTypeNames[TypeInteger], BuiltinTypeNames[dat])
	}

	_, err = d.Calculate(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}

	f := Attribute{
		id: "test-id-f",
		t:  TypeFloat}
	d = MakeAttributeDesignator(f)
	dat = d.GetResultType()
	if dat != TypeFloat {
		t.Errorf("Expected %q type but got %q", BuiltinTypeNames[TypeFloat], BuiltinTypeNames[dat])
	}

	_, err = d.Calculate(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	}
}
