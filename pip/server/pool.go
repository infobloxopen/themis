package server

type pool struct {
	ch chan []byte
}

func makePool(n, c int) pool {
	ch := make(chan []byte, n)
	for len(ch) < cap(ch) {
		ch <- make([]byte, 0, c)
	}

	return pool{
		ch: ch,
	}
}

func (p pool) get() []byte {
	return <-p.ch
}

func (p pool) put(b []byte) {
	p.ch <- b[:0]
}
