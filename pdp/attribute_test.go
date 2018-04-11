package pdp

import "testing"

func TestTypes(t *testing.T) {
	if typesTotal != len(BuiltinTypeNames) {
		t.Errorf("Expected total number of types to be equal to number of type names "+
			"but got typesTotal = %d and len(BuiltinTypeNames) = %d", typesTotal, len(BuiltinTypeNames))
	}
}

func TestAttribute(t *testing.T) {
	a := MakeAttribute("test", TypeString)
	if a.id != "test" {
		t.Errorf("Expected \"test\" as attribute id but got %q", a.id)
	}

	at := a.GetType()
	if at != TypeString {
		t.Errorf("Expected %q as attribute type but got %q", BuiltinTypeNames[TypeString], BuiltinTypeNames[at])
	}
}
