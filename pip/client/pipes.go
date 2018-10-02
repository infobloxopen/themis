package client

type pipes struct {
	idx chan int
	p   []pipe
}

func makePipes(n int) pipes {
	idx := make(chan int, n)
	for len(idx) < cap(idx) {
		idx <- len(idx)
	}

	p := make([]pipe, n)
	for i := range p {
		p[i] = makePipe()
	}

	return pipes{
		idx: idx,
		p:   p,
	}
}

func (p pipes) alloc() (int, pipe) {
	i := <-p.idx
	out := p.p[i]
	out.clean()
	return i, out
}

func (p pipes) free(i int) {
	p.idx <- i
}

func (p pipes) putError(i int, err error) {
	p.p[i].putError(err)
}

func (p pipes) putBytes(i int, b []byte) {
	p.p[i].putBytes(b)
}
