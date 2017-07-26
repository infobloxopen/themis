package pdp

import (
	"testing"

	"github.com/infobloxopen/go-trees/strtree"
)

func TestSortSetOfStrings(t *testing.T) {
	list := sortSetOfStrings(newStrTree("First", "Second", "Third"))

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

func newStrTree(args ...string) *strtree.Tree {
	t := strtree.NewTree()
	for i, s := range args {
		t.InplaceInsert(s, i)
	}

	return t
}
