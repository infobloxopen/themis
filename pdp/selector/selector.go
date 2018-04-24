package selector

import (
	"fmt"

	"github.com/infobloxopen/themis/pdp"
)

// MakeSelector returns new selector with given parameters
func MakeSelector(proto string, uri string, path []pdp.Expression, t pdp.Type) (pdp.Expression, error) {
	var ret pdp.Expression
	selector, ok := pdp.GetSelector(proto)
	if !ok {
		return ret, fmt.Errorf("unsupported selector scheme %s", proto)
	}
	if !selector.Enabled() {
		return ret, fmt.Errorf("selector scheme %s is not enabled", proto)
	}
	return selector.SelectorFunc(uri, path, t)
}
