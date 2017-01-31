package pdp

import (
	"golang.org/x/net/idna"
	"golang.org/x/text/unicode/norm"
	"strings"
)

type SetOfSubdomains struct {
	Final bool
	Sub   map[string]*SetOfSubdomains
}

func AdjustDomainName(s string) (string, error) {
	return idna.ToASCII(strings.ToLower(norm.NFC.String(s)))
}

func (s *SetOfSubdomains) addToSetOfDomains(d string) {
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
			nextNode = &SetOfSubdomains{false, make(map[string]*SetOfSubdomains)}
			node.Sub[label] = nextNode
		}

		node = nextNode
	}

	node.Final = true
}

func (s SetOfSubdomains) Contains(d string) bool {
	if s.Final {
		return true
	}

	if len(d) < 1 {
		return false
	}

	labels := strings.Split(d, ".")

	node := &s
	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]

		nextNode, ok := node.Sub[label]
		if !ok {
			return false
		}

		if nextNode.Final {
			return true
		}

		node = nextNode
	}

	return false
}
