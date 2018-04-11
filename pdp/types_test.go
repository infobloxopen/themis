package pdp

import "testing"

func TestTypes(t *testing.T) {
	if typesTotal != len(BuiltinTypeNames) {
		t.Errorf("Expected total number of types to be equal to number of type names "+
			"but got typesTotal = %d and len(BuiltinTypeNames) = %d", typesTotal, len(BuiltinTypeNames))
	}
}
