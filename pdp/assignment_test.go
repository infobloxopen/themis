package pdp

import (
	"testing"

	"github.com/infobloxopen/go-trees/domaintree"
)

func TestAttributeAssignmentExpression(t *testing.T) {
	ctx, err := NewContext(nil, 0, nil)
	if err != nil {
		t.Fatalf("Expected context but got error %s", err)
	}

	expect := "test-value"
	v := MakeStringValue(expect)
	a := Attribute{
		id: "test-id",
		t:  TypeString}

	ae := MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err := ae.Serialize(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if id != a.id || tName != BuiltinTypeKeys[a.t] || s != expect {
		t.Errorf("Expected %q, %q, %q but got %q, %q, %q", a.id, BuiltinTypeKeys[a.t], expect, id, tName, s)
	}

	dv := MakeDomainValue(domaintree.WireDomainNameLower("\x07example\x03com\x00"))
	v = MakeStringValue(expect)
	e := makeFunctionStringEqual(v, dv)
	a = Attribute{
		id: "test-id",
		t:  TypeBoolean}

	ae = MakeAttributeAssignmentExpression(a, e)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError error but got %T (%s)", err, err)
	}

	expect = "test-value"
	v = MakeStringValue(expect)
	a = Attribute{
		id: "test-id",
		t:  TypeBoolean}
	ae = MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*assignmentTypeMismatch); !ok {
		t.Errorf("Expected *ssignmentTypeMismatch error but got %T (%s)", err, err)
	}

	fv := MakeFloatValue(2.718282)
	v = MakeStringValue(expect)
	e = makeFunctionStringEqual(v, fv)
	a = Attribute{
		id: "test-id",
		t:  TypeInteger}

	ae = MakeAttributeAssignmentExpression(a, e)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError error but got %T (%s)", err, err)
	}

	v = MakeFloatValue(1234.567)
	a = Attribute{
		id: "test-id",
		t:  TypeInteger}
	ae = MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*assignmentTypeMismatch); !ok {
		t.Errorf("Expected *ssignmentTypeMismatch error but got %T (%s)", err, err)
	}

	iv := MakeIntegerValue(45678)
	v = MakeStringValue(expect)
	e = makeFunctionStringEqual(v, iv)
	a = Attribute{
		id: "test-id",
		t:  TypeFloat}

	ae = MakeAttributeAssignmentExpression(a, e)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*attributeValueTypeError); !ok {
		t.Errorf("Expected *attributeValueTypeError error but got %T (%s)", err, err)
	}

	expect = "45679.23"
	iv = MakeIntegerValue(45678)
	fv = MakeFloatValue(1.23)
	e = makeFunctionFloatAdd(fv, iv)
	a = Attribute{
		id: "test-id",
		t:  TypeFloat}

	ae = MakeAttributeAssignmentExpression(a, e)
	id, tName, s, err = ae.Serialize(ctx)
	if err != nil {
		t.Errorf("Expected no error but got %s", err)
	} else if id != a.id || tName != BuiltinTypeKeys[a.t] || s != expect {
		t.Errorf("Expected %q, %q, %q but got %q, %q, %q", a.id, BuiltinTypeKeys[a.t], expect, id, tName, s)
	}

	v = MakeIntegerValue(12345)
	a = Attribute{
		id: "test-id",
		t:  TypeFloat}
	ae = MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*assignmentTypeMismatch); !ok {
		t.Errorf("Expected *ssignmentTypeMismatch error but got %T (%s)", err, err)
	}

	v = UndefinedValue
	a = Attribute{
		id: "test-id",
		t:  TypeUndefined}
	ae = MakeAttributeAssignmentExpression(a, v)
	id, tName, s, err = ae.Serialize(ctx)
	if err == nil {
		t.Errorf("Expected error but got %q, %q, %q", id, tName, s)
	} else if _, ok := err.(*invalidTypeSerializationError); !ok {
		t.Errorf("Expected *invalidTypeSerializationError error but got %T (%s)", err, err)
	}
}
