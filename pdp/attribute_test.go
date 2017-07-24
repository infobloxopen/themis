package pdp

import "testing"

func TestAttribute(t *testing.T) {
	a := attribute{id: "test", t: typeString}

	err := a.newMissingError()
	if err == nil {
		t.Errorf("Expected error but got nothing")
	} else {
		_, ok := err.(*missingAttributeError)
		if !ok {
			t.Errorf("Expected *missingAttributeError but got %T", err)
		}
	}
}

func assertStringValue(v attributeValue, err error, e string, desc string, t *testing.T) {
	if err != nil {
		t.Errorf("Expected no error for %s but got %s", desc, err)
		return
	}

	s, err := v.str()
	if err != nil {
		t.Errorf("Expected string value for %s but got error \"%s\"", desc, err)
	}

	if s != e {
		t.Errorf("Expected %q for %s but got %q", e, desc, s)
	}
}
