package pdp

import (
	"fmt"
	"strings"
	"testing"
)

func TestDomainTreeInsert(t *testing.T) {
	tree := NewSetOfSubdomains()

	tree.insert("test.example.com", 0)
	assertDomainTreeIntValue(tree, "test.example.com", 0, t)
	assertDomainTreeIntValue(tree, "sub.test.example.com", 0, t)
	assertDomainTreeNoValue(tree, "example.com", t)

	tree.insert("test.com", 1)
	assertDomainTreeIntValue(tree, "www.test.com", 1, t)

	tree.insert("", 2)
	assertDomainTreeIntValue(tree, "example.com", 2, t)
	assertDomainTreeIntValue(tree, "example.net", 2, t)
	assertDomainTreeIntValue(tree, "test.example.com", 0, t)
}

func TestDomainTreeContains(t *testing.T) {
	tree := NewSetOfSubdomains()

	tree.insert("test.example.com", 0)
	if !tree.Contains("test.example.com") {
		t.Errorf("Expected some value at \"%s\" but got nothing")
	}
}

func TestDomainTreeIterate(t *testing.T) {
	tree := NewSetOfSubdomains()

	tree.insert("test.example.com", 0)
	tree.insert("example.com", 1)
	tree.insert("test.net", 2)
	tree.insert("www.example.biz", 2)

	domains := make(map[string]int)
	for item := range tree.Iterate() {
		i, ok := item.Leaf.(int)
		if !ok {
			t.Errorf("Expected only integer values but got %T (%#v) at \"%s\"", item.Leaf, item.Leaf, item.Domain)
		}

		domains[item.Domain] = i
	}

	assertMapStringIntEqual(domains, map[string]int{
		"test.example.com": 0,
		"example.com": 1,
		"test.net": 2,
		"www.example.biz": 2}, t)
}

func TestDomainTreeAdjustDomainName(t *testing.T) {
	raw := "example.com"
	domain, err := AdjustDomainName(raw)
	if err != nil {
		t.Errorf("Don't expect error for \"%s\" adjustment but got %s", raw, err)
	} else {
		if domain != raw {
			t.Errorf("Expected ajusted domain \"%s\" but got \"%s\"", raw, domain)
		}
	}

	raw = "\u043f\u0440\u0438\u043c\u0435\u0440.\u0440\u0444"
	conv := "xn--e1afmkfd.xn--p1ai"
	domain, err = AdjustDomainName(raw)
	if err != nil {
		t.Errorf("Don't expect error for \"%s\" adjustment but got %s", raw, err)
	} else {
		if domain != conv {
			t.Errorf("Expected ajusted domain \"%s\" but got \"%s\"", conv, domain)
		}
	}

	raw = "xn---"
	domain, err = AdjustDomainName(raw)
	if err == nil {
		t.Errorf("Expected error for domain \"%s\" but got converted to \"%s\"", raw, domain)
	}
}

func assertDomainTreeNoValue(tree *SetOfSubdomains, key string, t *testing.T) {
	v, ok := tree.Get(key)
	if ok {
		t.Errorf("Expected no value at \"%s\" but got %#v", key, v)
	}
}

func assertDomainTreeIntValue(tree *SetOfSubdomains, key string, i int, t *testing.T) {
	v, ok := tree.Get(key)
	if !ok {
		t.Errorf("Expected some value at \"%s\" but got nothing (%#v)", key, v)
		return
	}

	j, ok := v.(int)
	if !ok {
		t.Errorf("Expected integer value at \"%s\" but got %T (%#v)", key, v, v)
		return
	}

	if j != i {
		t.Errorf("Expected %d at \"%s\" but got %d", i, key, j)
	}
}

func assertMapStringIntEqual(m, e map[string]int, t *testing.T) {
	different := []string{}
	excess := []string{}
	for k, v := range m {
		ev, ok := e[k]
		if !ok {
			excess = append(excess, fmt.Sprintf("\t%s: %d", k, v))
		} else {
			if v != ev {
				different = append(different, fmt.Sprintf("\t%s: expected %d but got %d", k, ev, v))
			}
		}
	}

	absent := []string{}
	for k, v := range e {
		_, ok := m[k]
		if !ok {
			absent = append(absent, fmt.Sprintf("\t%s: %d", k, v))
		}
	}

	if len(different) > 0 || len(excess) > 0 || len(absent) > 0 {
		lines := []string{"Maps are different:"}

		if len(different) > 0 {
			lines = append(lines, "Different:")
			lines = append(lines, different...)
		}

		if len(excess) > 0 {
			lines = append(lines, "Excess:")
			lines = append(lines, excess...)
		}

		if len(absent) > 0 {
			lines = append(lines, "Absent:")
			lines = append(lines, absent...)
		}

		t.Errorf("%s\n", strings.Join(lines, "\n\t"))
	}
}
