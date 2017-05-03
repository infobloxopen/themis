package bitradix

type node32 struct {
	*Radix32
	branch int // -1 root, 0 left branch, 1 right branch
}
type queue32 []*node32

type node64 struct {
	*Radix64
	branch int
}
type queue64 []*node64

// Push adds a node32 to the queue.
func (q *queue32) Push(n *node32) {
	*q = append(*q, n)
}

// Pop removes and returns a node from the queue in first to last order.
func (q *queue32) Pop() *node32 {
	lq := len(*q)
	if lq == 0 {
		return nil
	}
	n := (*q)[0]
	switch lq {
	case 1:
		*q = (*q)[:0]
	default:
		*q = (*q)[1:lq]
	}
	return n
}

func (q *queue64) Push(n *node64) {
	*q = append(*q, n)
}

func (q *queue64) Pop() *node64 {
	lq := len(*q)
	if lq == 0 {
		return nil
	}
	n := (*q)[0]
	switch lq {
	case 1:
		*q = (*q)[:0]
	default:
		*q = (*q)[1:lq]
	}
	return n
}
