package pdp

import (
	log "github.com/sirupsen/logrus"
)

type Selector interface {
	Enabled() bool
	SelectorFunc(string, []Expression, int) (Expression, error)
}

var SelectorMap = make(map[string]Selector)

func RegisterSelector(name string, s Selector) {
	log.Debugf("Register %s selector", name)
	SelectorMap[name] = s
}
