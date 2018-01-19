package selector

import (
	"fmt"

	"github.com/infobloxopen/themis/pdp"
)

// MakeSelector returns new selector with given parameters
func MakeSelector(proto string, uri string, path []pdp.Expression, t int) (pdp.Expression, error) {
	var ret pdp.Expression
	selectorFunc, ok := pdp.SelectorMap[proto]
	if !ok {
		return ret, fmt.Errorf("Unsupported selector scheme %s", proto)
	}
	return selectorFunc(uri, path, t)
}
