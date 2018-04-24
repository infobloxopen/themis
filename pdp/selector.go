package pdp

type Selector interface {
	Enabled() bool
	Initialize()
	SelectorFunc(string, []Expression, Type) (Expression, error)
}

var selectorMap = make(map[string]Selector)

func RegisterSelector(name string, s Selector) {
	selectorMap[name] = s
}

func GetSelector(name string) (Selector, bool) {
	s, ok := selectorMap[name]
	return s, ok
}

func InitializeSelectors() {
	for _, e := range selectorMap {
		if e.Enabled() {
			e.Initialize()
		}
	}
}
