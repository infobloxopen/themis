package pdp

import (
	"testing"
)

func TestPIPSelector(t *testing.T) {
	sel := PIPSelector{
		service:   "10.0.0.2",
		queryType: "content-category",
		path: []Expression{
			MakeDomainValue("dodgysite.co.uk")},
		t: TypeString}

	c := &Context{}
	v, err := sel.calculate(c)
	assertStringValue(v, err, "Naughty", "simple PIP selector", t)
}
