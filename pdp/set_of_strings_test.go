package pdp

import "testing"

func TestSortSetOfStrings(t *testing.T) {
	list := SortSetOfStrings(newStrTree("First", "Second", "Third"))

	expected := []string{"First", "Second", "Third"}
	if len(list) != len(expected) {
		t.Fatalf("Expected %#v but got %#v", expected, list)
	}

	for i, item := range list {
		if item != expected[i] {
			t.Fatalf("Expected %#v but got %#v", expected, list)
		}
	}
}
