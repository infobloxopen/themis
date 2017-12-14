package local

import (
	"testing"

	"github.com/infobloxopen/go-trees/strtree"
	"github.com/infobloxopen/go-trees/strtrie"
	"github.com/infobloxopen/themis/pdp"
)

func TestLocalSelector(t *testing.T) {
	csmt := strtrie.NewPrefixTrie()
	csmt = csmt.Insert("test-key", "test-value")
	csm := pdp.MakeContentStringMap(csmt)
	cit := strtree.NewTree()
	cit.InplaceInsert("test-item", pdp.MakeContentMappingItem("test-item", pdp.TypeString, []int{pdp.TypeString}, csm))

	c, err := pdp.NewContext(
		pdp.NewLocalContentStorage(
			[]*pdp.LocalContent{
				pdp.NewLocalContent(
					"test-content",
					nil,
					[]*pdp.ContentItem{
						pdp.MakeContentMappingItem(
							"test-item",
							pdp.TypeString,
							[]int{pdp.TypeString},
							csm,
						),
					},
				),
			},
		),
		0,
		nil,
	)
	if err != nil {
		t.Fatalf("can't create context: %s", err)
	}

	sel := LocalSelector{
		content: "test-content",
		item:    "test-item",
		path: []pdp.Expression{
			pdp.MakeStringValue("test-key")},
		t: pdp.TypeString}
	v, err := sel.Calculate(c)
	assertStringValue(v, err, "test-value", "simple string selector", t)
}

func assertStringValue(v pdp.AttributeValue, err error, e string, desc string, t *testing.T) {
	if err != nil {
		t.Errorf("Expected no error for %s but got %s", desc, err)
		return
	}

	vt := v.GetResultType()
	if vt != pdp.TypeString {
		t.Errorf("Expected string for %s but got \"%s\"", desc, pdp.TypeNames[vt])
		return
	}

	s, err := v.Serialize()
	if err != nil {
		t.Errorf("Expected string value for %s but got error \"%s\"", desc, err)
	}

	if s != e {
		t.Errorf("Expected %q for %s but got %q", e, desc, s)
	}
}
