package client

type pipe chan response

func makePipe() pipe {
	return make(chan response, 1)
}

func (p pipe) clean() {
	for {
		select {
		default:
			return

		case <-p:
		}
	}
}

func (p pipe) get() ([]byte, error) {
	r := <-p
	return r.b, r.err
}

func (p pipe) putError(err error) {
	select {
	default:
	case p <- response{err: err}:
	}
}

func (p pipe) putBytes(b []byte) {
	select {
	default:
	case p <- response{b: b}:
	}
}
