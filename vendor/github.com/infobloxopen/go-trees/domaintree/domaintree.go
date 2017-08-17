// Package domaintree implements radix tree data structure for domain names.
package domaintree

import (
	"strings"

	"github.com/infobloxopen/go-trees/strtree"
)

// Node is a radix tree for domain names.
type Node struct {
	branches *strtree.Tree

	hasValue bool
	value    interface{}
}

// Pair represents a key-value pair returned by Enumerate method.
type Pair struct {
	Key   string
	Value interface{}
}

// Insert puts value using given domain as a key. The method returns new tree (old one remains unaffected). Input name converted to ASCII lowercase according to RFC-4343 (by mapping A-Z to a-z) to perform case-insensitive comparison when getting data from the tree.
func (n *Node) Insert(d string, v interface{}) *Node {
	if n == nil {
		n = &Node{}
	} else {
		n = &Node{
			branches: n.branches,
			hasValue: n.hasValue,
			value:    n.value}
	}
	r := n

	labels := strings.Split(asciiLowercase(d), ".")
	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]

		item, ok := n.branches.Get(label)
		var next *Node
		if ok {
			next = item.(*Node)
			next = &Node{
				branches: next.branches,
				hasValue: next.hasValue,
				value:    next.value}
		} else {
			next = &Node{}
		}

		n.branches = n.branches.Insert(label, next)
		n = next
	}

	n.hasValue = true
	n.value = v

	return r
}

// InplaceInsert puts or replaces value using given domain as a key. The method inserts data directly to current tree so make sure you have exclusive access to it. Input name converted in the same way as for Insert.
func (n *Node) InplaceInsert(d string, v interface{}) {
	if n.branches == nil {
		n.branches = strtree.NewTree()
	}

	labels := strings.Split(asciiLowercase(d), ".")
	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]

		item, ok := n.branches.Get(label)
		if ok {
			n = item.(*Node)
		} else {
			next := &Node{branches: strtree.NewTree()}
			n.branches.InplaceInsert(label, next)
			n = next
		}
	}

	n.hasValue = true
	n.value = v
}

// Enumerate returns key-value pairs in given tree sorted by key first by top level domain label second by second level and so on.
func (n *Node) Enumerate() chan Pair {
	ch := make(chan Pair)

	go func() {
		defer close(ch)
		n.enumerate("", ch)
	}()

	return ch
}

// Get gets value for domain which is equal to domain in the tree or is a subdomain of existing domain.
func (n *Node) Get(d string) (interface{}, bool) {
	if n == nil {
		return nil, false
	}

	labels := strings.Split(d, ".")
	for i := len(labels) - 1; i >= 0; i-- {
		label := asciiLowercase(labels[i])

		item, ok := n.branches.Get(label)
		if !ok {
			break
		}

		n = item.(*Node)
	}

	return n.value, n.hasValue
}

// Delete removes current domain and all its subdomains if any. It returns new tree and flag if deletion indeed occurs.
func (n *Node) Delete(d string) (*Node, bool) {
	if n == nil {
		return nil, false
	}

	if len(d) <= 0 {
		if n.hasValue || !n.branches.IsEmpty() {
			return &Node{}, true
		}

		return n, false
	}

	return n.del(strings.Split(d, "."))
}

func (n *Node) enumerate(s string, ch chan Pair) {
	if n == nil {
		return
	}

	if n.hasValue {
		ch <- Pair{
			Key:   s,
			Value: n.value}
	}

	for item := range n.branches.Enumerate() {
		sub := item.Key
		if len(s) > 0 {
			sub += "." + s
		}
		node := item.Value.(*Node)

		node.enumerate(sub, ch)
	}
}

func (n *Node) del(labels []string) (*Node, bool) {
	last := len(labels) - 1
	label := asciiLowercase(labels[last])
	if last == 0 {
		branches, ok := n.branches.Delete(label)
		if ok {
			return &Node{
				branches: branches,
				hasValue: n.hasValue,
				value:    n.value}, true
		}

		return n, false
	}

	item, ok := n.branches.Get(label)
	if !ok {
		return n, false
	}

	next := item.(*Node)
	next, ok = next.del(labels[:last])
	if !ok {
		return n, false
	}

	if next.branches.IsEmpty() && !next.hasValue {
		branches, _ := n.branches.Delete(label)
		return &Node{
			branches: branches,
			hasValue: n.hasValue,
			value:    n.value}, true
	}

	return &Node{
		branches: n.branches.Insert(label, next),
		hasValue: n.hasValue,
		value:    n.value}, true
}

func asciiLowercase(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}, s)
}
