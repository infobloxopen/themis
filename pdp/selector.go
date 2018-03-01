package pdp

import (
	log "github.com/sirupsen/logrus"
)

type SelectorFunc func(string, []Expression, int) (Expression, error)

var SelectorMap = map[string]SelectorFunc{}

func RegisterSelector(name string, fn SelectorFunc) {
	log.Debugf("Register %s selector", name)
	SelectorMap[name] = fn
}
