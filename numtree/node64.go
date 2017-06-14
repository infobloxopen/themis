package numtree

import (
	"errors"
	"fmt"
)

const key64BitSize = 64

// ErrBits64OutOfRange is the error returned when number of significant bits out of interval [0, 64].
var (
	ErrBits64OutOfRange = errors.New("number of significant bits is out of range (expected 0 <= n < 64)")

	masks64 = []uint64{
		0x0000000000000000, 0x8000000000000000, 0xc000000000000000, 0xe000000000000000,
		0xf000000000000000, 0xf800000000000000, 0xfc00000000000000, 0xfe00000000000000,
		0xff00000000000000, 0xff80000000000000, 0xffc0000000000000, 0xffe0000000000000,
		0xfff0000000000000, 0xfff8000000000000, 0xfffc000000000000, 0xfffe000000000000,
		0xffff000000000000, 0xffff800000000000, 0xffffc00000000000, 0xffffe00000000000,
		0xfffff00000000000, 0xfffff80000000000, 0xfffffc0000000000, 0xfffffe0000000000,
		0xffffff0000000000, 0xffffff8000000000, 0xffffffc000000000, 0xffffffe000000000,
		0xfffffff000000000, 0xfffffff800000000, 0xfffffffc00000000, 0xfffffffe00000000,
		0xffffffff00000000, 0xffffffff80000000, 0xffffffffc0000000, 0xffffffffe0000000,
		0xfffffffff0000000, 0xfffffffff8000000, 0xfffffffffc000000, 0xfffffffffe000000,
		0xffffffffff000000, 0xffffffffff800000, 0xffffffffffc00000, 0xffffffffffe00000,
		0xfffffffffff00000, 0xfffffffffff80000, 0xfffffffffffc0000, 0xfffffffffffe0000,
		0xffffffffffff0000, 0xffffffffffff8000, 0xffffffffffffc000, 0xffffffffffffe000,
		0xfffffffffffff000, 0xfffffffffffff800, 0xfffffffffffffc00, 0xfffffffffffffe00,
		0xffffffffffffff00, 0xffffffffffffff80, 0xffffffffffffffc0, 0xffffffffffffffe0,
		0xfffffffffffffff0, 0xfffffffffffffff8, 0xfffffffffffffffc, 0xfffffffffffffffe,
		0xffffffffffffffff}
)

// Node64 is an element of radix tree with 64-bit unsigned integer as a key.
type Node64 struct {
	// Key stores key for current node.
	Key uint64
	// Bits is a number of significant bits in Key.
	Bits uint8
	// Leaf indicates if the node is leaf node and contains any data in Value.
	Leaf bool
	// Value contains data associated with key.
	Value interface{}

	chld [2]*Node64
}

// Dot dumps tree to Graphviz .dot format
func (n *Node64) Dot() string {
	body := ""

	i := 0
	queue := []*Node64{n}
	for len(queue) > 0 {
		c := queue[0]
		body += fmt.Sprintf("N%d %s\n", i, c.dotString())

		if c != nil && (c.chld[0] != nil || c.chld[1] != nil) {
			body += fmt.Sprintf("N%d -> { N%d N%d }\n", i, i+len(queue), i+len(queue)+1)
			queue = append(append(queue, c.chld[0]), c.chld[1])
		}

		queue = queue[1:]
		i++
	}

	return "digraph d {\n" + body + "}\n"
}

// Insert puts new leaf to radix tree and returns pointer to new root. The method uses copy on write strategy so old root doesn't see the change.
func (n *Node64) Insert(key uint64, bits int, value interface{}) (*Node64, error) {
	if bits < 0 || bits > key64BitSize {
		return nil, ErrBits64OutOfRange
	}

	return n.insert(newNode64(key&masks64[bits], uint8(bits), true, value)), nil
}

// Get locates node which key is equal to or "contains" the key passed as argument.
func (n *Node64) Get(key uint64, bits int) (interface{}, bool) {
	if bits < 0 {
		bits = 0
	} else if bits > key64BitSize {
		bits = key64BitSize
	}

	r := n.get(key, uint8(bits))
	if r == nil {
		return nil, false
	}

	return r.Value, true
}

func (n *Node64) dotString() string {
	if n == nil {
		return "[label=\"nil\"]"
	}

	if n.Leaf {
		v := fmt.Sprintf("%q", fmt.Sprintf("%#v", n.Value))
		return fmt.Sprintf("[label=\"k: %016x, b: %d, v: \\\"%s\\\"\"]", n.Key, n.Bits, v[1:len(v)-1])
	}

	return fmt.Sprintf("[label=\"k: %016x, b: %d\"]", n.Key, n.Bits)
}

func (n *Node64) insert(c *Node64) *Node64 {
	if n == nil {
		return c
	}

	bits := clz64((n.Key ^ c.Key) | ^masks64[n.Bits] | ^masks64[c.Bits])
	if bits < n.Bits {
		branch := (n.Key >> (key64BitSize - 1 - bits)) & 1
		if bits == c.Bits {
			c.chld[branch] = n
			return c
		}

		m := newNode64(c.Key&masks64[bits], bits, false, nil)
		m.chld[branch] = n
		m.chld[1-branch] = c

		return m
	}

	if c.Bits == n.Bits {
		c.chld = n.chld
		return c
	}

	m := newNode64(n.Key, n.Bits, n.Leaf, n.Value)
	m.chld = n.chld

	branch := (c.Key >> (key64BitSize - 1 - bits)) & 1
	m.chld[branch] = m.chld[branch].insert(c)

	return m
}

func (n *Node64) get(key uint64, bits uint8) *Node64 {
	if n == nil || n.Bits > bits {
		return nil
	}

	mask := masks64[n.Bits]
	if n.Bits == bits {
		if n.Leaf && (n.Key^key)&mask == 0 {
			return n
		}

		return nil
	}

	c := n.chld[(key>>(key64BitSize-1-n.Bits))&1]
	if c != nil {
		r := c.get(key, bits)
		if r != nil {
			return r
		}
	}

	if n.Leaf && (n.Key^key)&mask == 0 {
		return n
	}

	return nil
}

func newNode64(key uint64, bits uint8, leaf bool, value interface{}) *Node64 {
	return &Node64{
		Key:   key,
		Bits:  bits,
		Leaf:  leaf,
		Value: value}
}

// clz64 counts leading zeroes in unsigned 64-bit integer using binary search combined with table lookup for last 4 bits.
func clz64(x uint64) uint8 {
	var n uint8

	if x&0xffffffff00000000 == 0 {
		n += 32
		x <<= 32
	}

	if x&0xffff000000000000 == 0 {
		n += 16
		x <<= 16
	}

	if x&0xff00000000000000 == 0 {
		n += 8
		x <<= 8
	}

	if x&0xf000000000000000 == 0 {
		n += 4
		x <<= 4
	}

	return n + clzTable[x>>(key64BitSize-4)]
}
