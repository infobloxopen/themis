package bitradix

// Radix64 implements a radix tree with an uint64 as its key.
type Radix64 struct {
	branch [2]*Radix64 // branch[0] is left branch for 0, and branch[1] the right for 1
	parent *Radix64
	key    uint64      // the key under which this value is stored
	bits   int         // the number of significant bits, if 0 the key has not been set.
	Value  interface{} // The value stored.
}

func New64() *Radix64 {
	// It gets two branches by default
	return &Radix64{[2]*Radix64{
		&Radix64{[2]*Radix64{nil, nil}, nil, 0, 0, nil},
		&Radix64{[2]*Radix64{nil, nil}, nil, 0, 0, nil},
	}, nil, 0, 0, nil}
}

func (r *Radix64) Key() uint64 {
	return r.key
}

func (r *Radix64) Bits() int {
	return r.bits
}

func (r *Radix64) Leaf() bool {
	return r.branch[0] == nil && r.branch[1] == nil
}

func (r *Radix64) Insert(n uint64, bits int, v interface{}) *Radix64 {
	if r.parent != nil {
		panic("bitradix: not the root node")
	}
	return r.insert(n, bits, v, bitSize32-1)
}

func (r *Radix64) Remove(n uint64, bits int) *Radix64 {
	if r.parent != nil {
		panic("bitradix: not the root node")
	}
	return r.remove(n, bits, bitSize32-1)
}

func (r *Radix64) Find(n uint64, bits int) *Radix64 {
	if r.parent != nil {
		panic("bitradix: not the root node")
	}
	return r.find(n, bits, bitSize32-1, nil)
}

func (r *Radix64) Do(f func(*Radix64, int)) {
	q := make(queue64, 0)

	q.Push(&node64{r, -1})
	x := q.Pop()
	for x != nil {
		f(x.Radix64, x.branch)
		for i, b := range x.Radix64.branch {
			if b != nil {
				q.Push(&node64{b, i})
			}
		}
		x = q.Pop()
	}
}

func (r *Radix64) insert(n uint64, bits int, v interface{}, bit int) *Radix64 {
	switch r.Leaf() {
	case false: // Non-leaf node, one or two branches, possibly a key
		if bit < 0 {
			panic("bitradix: bit index smaller than zero")
		}
		bnew := bitK64(n, bit)
		if r.bits == 0 && bits == bitSize32-bit { // I should be put here
			r.set(n, bits, v)
			return r
		}
		if r.bits > 0 && bits == bitSize32-bit {
			bcur := bitK64(r.key, bit)
			if r.bits > bits {
				b1 := r.bits
				n1 := r.key
				v1 := r.Value
				r.set(n, bits, v)
				if r.branch[bcur] == nil {
					r.branch[bcur] = r.new()
				}
				r.branch[bcur].insert(n1, b1, v1, bit-1)
				return r
			}
		}
		if r.branch[bnew] == nil {
			r.branch[bnew] = r.new()
		}
		return r.branch[bnew].insert(n, bits, v, bit-1)
	case true: // External node, (optional) key, no branches
		if r.bits == 0 || r.key == n { // nothing here yet, put something in, or equal keys
			r.set(n, bits, v)
			return r
		}
		if bit < 0 {
			panic("bitradix: bit index smaller than zero")
		}
		bcur := bitK64(r.key, bit)
		bnew := bitK64(n, bit)
		if bcur == bnew {
			r.branch[bcur] = r.new()
			if r.bits > 0 && (bits == bitSize32-bit || bits < r.bits) {
				b1 := r.bits
				n1 := r.key
				v1 := r.Value
				r.set(n, bits, v)
				r.branch[bnew].insert(n1, b1, v1, bit-1)
				return r
			}
			if r.bits > 0 && bits >= r.bits {
				// current key can not be put further down, leave it
				// but continue
				return r.branch[bnew].insert(n, bits, v, bit-1)
			}
			// fill this node, with the current key - and call ourselves
			r.branch[bcur].set(r.key, r.bits, r.Value)
			r.clear()
			return r.branch[bnew].insert(n, bits, v, bit-1)
		}
		// not equal, keep current node, and branch off in child
		r.branch[bcur] = r.new()
		// fill this node, with the current key - and call ourselves
		r.branch[bcur].set(r.key, r.bits, r.Value)
		r.clear()
		r.branch[bnew] = r.new()
		return r.branch[bnew].insert(n, bits, v, bit-1)
	}
	panic("bitradix: not reached")
}

func (r *Radix64) remove(n uint64, bits, bit int) *Radix64 {
	if r.bits > 0 && r.bits == bits {
		// possible hit
		mask := uint64(mask64 << (bitSize32 - uint(r.bits)))
		if r.key&mask == n&mask {
			// save r in r1
			r1 := &Radix64{[2]*Radix64{nil, nil}, nil, r.key, r.bits, r.Value}
			r.prune(true)
			return r1
		}
	}
	k := bitK64(n, bit)
	if r.Leaf() || r.branch[k] == nil { // dead end
		return nil
	}
	return r.branch[bitK64(n, bit)].remove(n, bits, bit-1)
}

func (r *Radix64) prune(b bool) {
	if b {
		if r.parent == nil {
			r.clear()
			return
		}
		// we are a node, we have a parent, so the parent is a non-leaf node
		if r.parent.branch[0] == r {
			// kill that branch
			r.parent.branch[0] = nil
		}
		if r.parent.branch[1] == r {
			r.parent.branch[1] = nil
		}
		r.parent.prune(false)
		return
	}
	if r == nil {
		return
	}
	if r.bits != 0 {
		// fun stops
		return
	}
	// Does I have one or two childeren, if one, move my self up one node
	// Also the child must be a leaf node!
	b0 := r.branch[0]
	b1 := r.branch[1]
	if b0 != nil && b1 != nil {
		// two branches, we cannot replace ourselves with a child
		return
	}
	if b0 != nil {
		if !b0.Leaf() {
			return
		}
		// move b0 into this node	
		r.set(b0.key, b0.bits, b0.Value)
		r.branch[0] = b0.branch[0]
		r.branch[1] = b0.branch[1]
	}
	if b1 != nil {
		if !b1.Leaf() {
			return
		}
		// move b1 into this node
		r.set(b1.key, b1.bits, b1.Value)
		r.branch[0] = b1.branch[0]
		r.branch[1] = b1.branch[1]
	}
	r.parent.prune(false)
}

func (r *Radix64) find(n uint64, bits, bit int, last *Radix64) *Radix64 {
	switch r.Leaf() {
	case false:
		// A prefix that is matching (BETTER MATCHING)
		mask := uint64(mask64 << (bitSize32 - uint(r.bits)))
		if r.bits > 0 && r.key&mask == n&mask {
			//			fmt.Printf("Setting last to %d %s\n", r.key, r.Value)
			if last == nil {
				last = r
			} else {
				// Only when bigger
				if r.bits >= last.bits {
					last = r
				}
			}
		}
		if r.bits == bits && r.key&mask == n&mask {
			// our key
			return r
		}

		k := bitK64(n, bit)
		if r.branch[k] == nil {
			return last // REALLY?
		}
		return r.branch[k].find(n, bits, bit-1, last)
	case true:
		// It this our key...!?
		mask := uint64(mask64 << (bitSize32 - uint(r.bits)))
		if r.key&mask == n&mask {
			return r
		}
		return last
	}
	panic("bitradix: not reached")
}

func (r *Radix64) new() *Radix64 {
	return &Radix64{[2]*Radix64{nil, nil}, r, 0, 0, nil}
}

func (r *Radix64) set(key uint64, bits int, value interface{}) {
	r.key = key
	r.bits = bits
	r.Value = value
}

func (r *Radix64) clear() {
	r.key = 0
	r.bits = 0
	r.Value = nil
}

func bitK64(n uint64, k int) byte {
	return byte((n & (1 << uint(k))) >> uint(k))
}
