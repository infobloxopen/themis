package pdp

import (
	"io"
)

// Marshaler is interface that wraps AST marshal methods.
type Marshaler interface {
	Marshal(out io.Writer) error
	Size() int // provide an estimation of bytesize of the encoded output
}

// CappedBuffer creates a limited buffer to reduce the number of dynamic
// resizing in ResponseWriter. We make the assumption that buffer cap is at
// least > than the average input message len
type CappedBuffer struct {
	dest io.Writer
	buf  []byte
}

// NewCappedBuffer creates a CappedBuffer of capacity bufcap
func NewCappedBuffer(resp io.Writer, bufcap int) *CappedBuffer {
	return &CappedBuffer{
		dest: resp,
		buf:  make([]byte, 0, bufcap),
	}
}

// Write implements io.Writer interface
func (resp *CappedBuffer) Write(input []byte) (int, error) {
	start := len(resp.buf)
	limit := cap(resp.buf)
	if start+len(input) > limit {
		// write buffer to dest
		n, err := resp.dest.Write(resp.buf)
		if err != nil {
			return n, err
		}
		// write all of data to dest
		ninput, err := resp.dest.Write(input)
		if err != nil {
			return n + ninput, err
		}
		// reset
		resp.buf = resp.buf[:0]
		return n + ninput, nil
	}
	return copy(resp.buf[start:], input), nil
}
