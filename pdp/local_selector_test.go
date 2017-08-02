package pdp

import (
	"testing"

	"github.com/infobloxopen/go-trees/strtree"
)

func TestLocalSelector(t *testing.T) {
	csmt := strtree.NewTree()
	csmt.InplaceInsert("test-key", "test-value")
	csm := contentStringMap{csmt}
	cit := strtree.NewTree()
	cit.InplaceInsert("test-item", ContentItem{r: csm, t: TypeString, k: []int{TypeString}})
	ct := strtree.NewTree()
	ct.InplaceInsert("test-content", cit)

	c := &Context{c: ct}

	sel := LocalSelector{
		content: "test-content",
		item:    "test-item",
		path: []Expression{
			MakeStringValue("test-key")},
		t: TypeString}
	v, err := sel.calculate(c)
	assertStringValue(v, err, "test-value", "simple string selector", t)
}
