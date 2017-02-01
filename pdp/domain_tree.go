package pdp

import (
	"golang.org/x/net/idna"
	"golang.org/x/text/unicode/norm"
	"strings"
)

type SetOfSubdomains struct {
	Final bool
	Leaf  interface{}
	Sub   map[string]*SetOfSubdomains
}

func AdjustDomainName(s string) (string, error) {
	return idna.ToASCII(strings.ToLower(norm.NFC.String(s)))
}

func (s *SetOfSubdomains) addToSetOfDomains(d string, v interface{}) {
	if len(d) < 1 {
		s.Final = true
		return
	}

	labels := strings.Split(d, ".")

	node := s
	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]

		nextNode, ok := node.Sub[label]
		if !ok {
			nextNode = &SetOfSubdomains{false, nil, make(map[string]*SetOfSubdomains)}
			node.Sub[label] = nextNode
		}

		node = nextNode
	}

	node.Final = true
	node.Leaf = v
}

func (s SetOfSubdomains) Get(d string) (interface{}, bool) {
	if s.Final {
		return s.Leaf, true
	}

	if len(d) < 1 {
		return nil, false
	}

	labels := strings.Split(d, ".")

	node := &s
	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]

		nextNode, ok := node.Sub[label]
		if !ok {
			return nil, false
		}

		if nextNode.Final {
			return nextNode.Leaf, true
		}

		node = nextNode
	}

	return nil, false
}

func (s SetOfSubdomains) Contains(d string) bool {
	_, ok := s.Get(d)
	return ok
}
