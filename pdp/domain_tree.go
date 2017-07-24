package pdp

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
)

type SetOfSubdomains struct {
	branches map[string]*SetOfSubdomains

	hasValue bool
	value    interface{}
}

type DomainLeafItem struct {
	Domain string
	Leaf   interface{}
}

var domainRegexp = regexp.MustCompile("^[-._a-z0-9]+$")

func AdjustDomainName(s string) (string, error) {
	tmp, err := idna.Punycode.ToASCII(s)
	if err != nil {
		return "", fmt.Errorf("Cannot convert domain [%s]", s)
	}
	ret := strings.ToLower(tmp)
	if !domainRegexp.MatchString(ret) {
		return "", fmt.Errorf("Cannot validate domain [%s]", s)
	}
	return ret, nil
}

func NewSetOfSubdomains() *SetOfSubdomains {
	return &SetOfSubdomains{branches: make(map[string]*SetOfSubdomains)}
}

func (s *SetOfSubdomains) insert(d string, v interface{}) {
	node := s

	if len(d) > 0 {
		labels := strings.Split(d, ".")

		for i := len(labels) - 1; i >= 0; i-- {
			label := labels[i]

			next, ok := node.branches[label]
			if !ok {
				next = NewSetOfSubdomains()
				node.branches[label] = next
			}

			node = next
		}
	}

	node.hasValue = true
	node.value = v
}

func (s SetOfSubdomains) Get(d string) (interface{}, bool) {
	node := &s
	hasValue := node.hasValue
	value := node.value

	if len(d) > 0 {
		labels := strings.Split(d, ".")

		for i := len(labels) - 1; i >= 0; i-- {
			label := labels[i]

			next, ok := node.branches[label]
			if !ok {
				break
			}

			node = next
			if node.hasValue {
				hasValue = true
				value = node.value
			}
		}
	}

	return value, hasValue
}

func (s SetOfSubdomains) Contains(d string) bool {
	_, ok := s.Get(d)
	return ok
}

func (s *SetOfSubdomains) iterate(path []string, ch chan DomainLeafItem) {
	if s.hasValue {
		name := make([]string, len(path))
		for i, item := range path {
			name[len(path)-i-1] = item
		}

		ch <- DomainLeafItem{strings.Join(name, "."), s.value}
	}

	for name, subdomain := range s.branches {
		subdomain.iterate(append(path, name), ch)
	}
}

func (s *SetOfSubdomains) Iterate() chan DomainLeafItem {
	ch := make(chan DomainLeafItem)

	go func() {
		defer close(ch)

		s.iterate([]string{}, ch)
	}()

	return ch
}
