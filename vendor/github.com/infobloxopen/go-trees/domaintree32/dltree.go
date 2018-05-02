package domaintree32

// !!!DON'T EDIT!!! Generated with infobloxopen/go-trees/etc from domaintree{{.bits}} with etc -s uint32 -d dtuintX.yaml -t ./domaintree\{\{.bits\}\}

import "github.com/infobloxopen/go-trees/domain"

type labelTree struct {
	root *node
}

type labelPair struct {
	Key   string
	Value *Node
}

type labelRawPair struct {
	Key   domain.Label
	Value *Node
}

func newLabelTree() *labelTree {
	return new(labelTree)
}

func (t *labelTree) insert(key string, value *Node) *labelTree {
	var (
		n *node
	)

	if t != nil {
		n = t.root
	}

	dl, _ := domain.MakeLabel(key)
	return &labelTree{root: n.insert(dl, value)}
}

func (t *labelTree) rawInsert(key []byte, value *Node) *labelTree {
	var (
		n *node
	)

	if t != nil {
		n = t.root
	}

	return &labelTree{root: n.insert(domain.Label(key), value)}
}

func (t *labelTree) inplaceInsert(key string, value *Node) {
	dl, _ := domain.MakeLabel(key)
	t.root = t.root.inplaceInsert(dl, value)
}

func (t *labelTree) rawInplaceInsert(key []byte, value *Node) {
	t.root = t.root.inplaceInsert(domain.Label(key), value)
}

func (t *labelTree) get(key string) (*Node, bool) {
	if t == nil {
		return nil, false
	}

	dl, _ := domain.MakeLabel(key)
	return t.root.get(dl)
}

func (t *labelTree) rawGet(key []byte) (*Node, bool) {
	if t == nil {
		return nil, false
	}

	return t.root.get(domain.Label(key))
}

func (t *labelTree) enumerate() chan labelPair {
	ch := make(chan labelPair)

	go func() {
		defer close(ch)

		if t == nil {
			return
		}

		t.root.enumerate(ch)
	}()

	return ch
}

func (t *labelTree) rawEnumerate() chan labelRawPair {
	ch := make(chan labelRawPair)

	go func() {
		defer close(ch)

		if t == nil {
			return
		}

		t.root.rawEnumerate(ch)
	}()

	return ch
}

func (t *labelTree) del(key string) (*labelTree, bool) {
	if t == nil {
		return nil, false
	}

	dl, _ := domain.MakeLabel(key)
	root, ok := t.root.del(dl)
	return &labelTree{root: root}, ok
}

func (t *labelTree) rawDel(key []byte) (*labelTree, bool) {
	if t == nil {
		return nil, false
	}

	root, ok := t.root.del(domain.Label(key))
	return &labelTree{root: root}, ok
}

func (t *labelTree) isEmpty() bool {
	return t == nil || t.root == nil
}

func (t *labelTree) dot() string {
	body := ""

	if t != nil {
		body = t.root.dot()
	}

	return "digraph d {\n" + body + "}\n"
}
