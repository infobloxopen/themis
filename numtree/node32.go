package numtree

import (
	"errors"
	"fmt"
)

const key32BitSize = 32

// ErrBits32OutOfRange is the error returned when number of significant bits out of interval [0, 32].
var (
	ErrBits32OutOfRange = errors.New("number of significant bits is out of range (expected 0 <= n < 32)")

	masks32 = []uint32{
		0x00000000, 0x80000000, 0xc0000000, 0xe0000000,
		0xf0000000, 0xf8000000, 0xfc000000, 0xfe000000,
		0xff000000, 0xff800000, 0xffc00000, 0xffe00000,
		0xfff00000, 0xfff80000, 0xfffc0000, 0xfffe0000,
		0xffff0000, 0xffff8000, 0xffffc000, 0xffffe000,
		0xfffff000, 0xfffff800, 0xfffffc00, 0xfffffe00,
		0xffffff00, 0xffffff80, 0xffffffc0, 0xffffffe0,
		0xfffffff0, 0xfffffff8, 0xfffffffc, 0xfffffffe,
		0xffffffff}
)

// Node32 is an element of radix tree with 32-bit unsigned integer as a key.
type Node32 struct {
	// Key stores key for current node.
	Key uint32
	// Bits is a number of significant bits in Key.
	Bits uint8
	// Leaf indicates if the node is leaf node and contains any data in Value.
	Leaf bool
	// Value contains data associated with key.
	Value interface{}

	chld [2]*Node32
}

// Dot dumps tree to Graphviz .dot format
func (n *Node32) Dot() string {
	body := ""

	// Iterate all nodes using breadth-first search algorithm.
	i := 0
	queue := []*Node32{n}
	for len(queue) > 0 {
		c := queue[0]
		body += fmt.Sprintf("N%d %s\n", i, c.dotString())
		if c != nil && (c.chld[0] != nil || c.chld[1] != nil) {
			// Children for current node if any always go to the end of the queue
			// so we can know their indices using current queue length.
			body += fmt.Sprintf("N%d -> { N%d N%d }\n", i, i+len(queue), i+len(queue)+1)
			queue = append(append(queue, c.chld[0]), c.chld[1])
		}

		queue = queue[1:]
		i++
	}

	return "digraph d {\n" + body + "}\n"
}

// Insert puts new leaf to radix tree and returns pointer to new root. The method uses copy on write strategy so old root doesn't see the change.
func (n *Node32) Insert(key uint32, bits int, value interface{}) (*Node32, error) {
	if bits < 0 || bits > key32BitSize {
		return nil, ErrBits32OutOfRange
	}

	return n.insert(newNode32(key&masks32[bits], uint8(bits), true, value)), nil
}

// Get locates node which key is equal to or "contains" the key passed as argument.
func (n *Node32) Get(key uint32, bits int) (interface{}, bool) {
	if bits < 0 {
		bits = 0
	} else if bits > key32BitSize {
		bits = key32BitSize
	}

	r := n.get(key&masks32[bits], uint8(bits))
	if r == nil {
		return nil, false
	}

	return r.Value, true
}

func (n *Node32) dotString() string {
	if n == nil {
		return "[label=\"nil\"]"
	}

	if n.Leaf {
		v := fmt.Sprintf("%q", fmt.Sprintf("%#v", n.Value))
		return fmt.Sprintf("[label=\"k: %08x, b: %d, v: \\\"%s\\\"\"]", n.Key, n.Bits, v[1:len(v)-1])
	}

	return fmt.Sprintf("[label=\"k: %08x, b: %d\"]", n.Key, n.Bits)
}

func (n *Node32) insert(c *Node32) *Node32 {
	if n == nil {
		return c
	}

	// Find number of common most significant bits (NCSB):
	// 1. xor operation puts zeroes at common bits;
	// 2. or masks put ones so that zeroes can't go after smaller number of significant bits (NSB)
	// 3. count of leading zeroes gives number of common bits
	bits := clz32((n.Key ^ c.Key) | ^masks32[n.Bits] | ^masks32[c.Bits])

	// There are three cases possible:
	// - NCSB less than number of significant bits (NSB) of current tree node:
	if bits < n.Bits {
		// (branch for current tree node is determined by a bit right after the last common bit)
		branch := (n.Key >> (key32BitSize - 1 - bits)) & 1

		// - NCSB equals to NSB of candidate node:
		if bits == c.Bits {
			// make new root from the candidate and put current node to one of its branch;
			c.chld[branch] = n
			return c
		}

		// - NCSB less than NSB of candidate node (it can't be greater because bits after NSB don't count):
		// make new root (non-leaf node)
		m := newNode32(c.Key&masks32[bits], bits, false, nil)
		// with current tree node at one of branches
		m.chld[branch] = n
		// and the candidate at the other.
		m.chld[1-branch] = c

		return m
	}

	// - keys are equal (NCSB not less than NSB of current tree node and both numbers are equal):
	if c.Bits == n.Bits {
		// replace current node with the candidate.
		c.chld = n.chld
		return c
	}

	// - current tree node contains candidate node:
	// make new root as a copy of current tree node;
	m := newNode32(n.Key, n.Bits, n.Leaf, n.Value)
	m.chld = n.chld

	// (branch for the candidate is determined by a bit right after the last common bit)
	branch := (c.Key >> (key32BitSize - 1 - bits)) & 1
	// insert it to correct branch.
	m.chld[branch] = m.chld[branch].insert(c)

	return m
}

func (n *Node32) get(key uint32, bits uint8) *Node32 {
	// If tree is empty or current key can't be contained in root node -
	if n == nil || n.Bits > bits {
		// report nothing.
		return nil
	}

	mask := masks32[n.Bits]
	// If NSB of current tree node is the same as key has -
	if n.Bits == bits {
		// return current node only if it contains data (leaf node) and masked keys are equal.
		if n.Leaf && (n.Key^key)&mask == 0 {
			return n
		}

		return nil
	}

	// Otherwise jump to branch by key bit right after NSB of current tree node
	c := n.chld[(key>>(key32BitSize-1-n.Bits))&1]
	if c != nil {
		// and check if child on the branch has anything.
		r := c.get(key, bits)
		if r != nil {
			return r
		}
	}

	// If nothing matches check if current node contains the key given.
	if n.Leaf && (n.Key^key)&mask == 0 {
		return n
	}

	return nil
}

func newNode32(key uint32, bits uint8, leaf bool, value interface{}) *Node32 {
	return &Node32{
		Key:   key,
		Bits:  bits,
		Leaf:  leaf,
		Value: value}
}

// clz32 counts leading zeroes in unsigned 32-bit integer using binary search combined with table lookup for last 4 bits.
func clz32(x uint32) uint8 {
	var n uint8

	if x&0xffff0000 == 0 {
		n = 16
		x <<= 16
	}

	if x&0xff000000 == 0 {
		n += 8
		x <<= 8
	}

	if x&0xf0000000 == 0 {
		n += 4
		x <<= 4
	}

	return n + clzTable[x>>(key32BitSize-4)]
}
