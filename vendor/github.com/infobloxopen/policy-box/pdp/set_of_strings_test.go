package pdp

import "testing"

func TestSortSetOfStrings(t *testing.T) {
	list := sortSetOfStrings(map[string]int{"Third": 2, "First": 0, "Second": 1})

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
