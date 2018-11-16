package pip

import (
	"github.com/infobloxopen/themis/pdp"
)

func init() {
	pdp.RegisterSelector(new(selector))
	pdp.RegisterSelector(new(selectorUnix))
	pdp.RegisterSelector(new(selectorK8s))
}
