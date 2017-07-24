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
	cit.InplaceInsert("test-item", contentItem{r: csm, t: typeString})
	ct := strtree.NewTree()
	ct.InplaceInsert("test-content", cit)

	c := &Context{c: ct}

	sel := localSelector{
		content: "test-content",
		item:    "test-item",
		path: []expression{
			makeStringValue("test-key")}}
	v, err := sel.calculate(c)
	assertStringValue(v, err, "test-value", "simple string selector", t)
}
