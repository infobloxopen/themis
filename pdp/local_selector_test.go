package pdp

import (
	"testing"

	"github.com/infobloxopen/go-trees/strtree"
)

func TestLocalSelector(t *testing.T) {
	csmt := strtree.NewTree()
	csmt.InplaceInsert("test-key", "test-value")
	csm := MakeContentStringMap(csmt)
	cit := strtree.NewTree()
	cit.InplaceInsert("test-item", MakeContentMappingItem("test-item", TypeString, []int{TypeString}, csm))
	ct := strtree.NewTree()
	ct.InplaceInsert("test-content", &LocalContent{id: "test-content", items: cit})

	c := &Context{c: &LocalContentStorage{r: ct}}

	sel := LocalSelector{
		content: "test-content",
		item:    "test-item",
		path: []Expression{
			MakeStringValue("test-key")},
		t: TypeString}
	v, err := sel.calculate(c)
	assertStringValue(v, err, "test-value", "simple string selector", t)
}
