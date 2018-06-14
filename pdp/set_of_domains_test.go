package pdp

import "testing"

func TestSortSetOfDomains(t *testing.T) {
	list := SortSetOfDomains(newDomainTree(
		makeTestDomain("example.com"),
		makeTestDomain("example.gov"),
		makeTestDomain("www.example.com"),
	))

	expected := []string{"example.com", "example.gov", "www.example.com"}
	if len(list) != len(expected) {
		t.Fatalf("Expected %#v but got %#v", expected, list)
	}

	for i, item := range list {
		if item != expected[i] {
			t.Fatalf("Expected %#v but got %#v", expected, list)
		}
	}
}
