package pdp

import (
	"io"
	"net/http"
)

// Marshaler is interface that wraps AST marshal methods.
type Marshaler interface {
	Marshal(out io.Writer) error
	Size() int // provide an estimation of bytesize of the encoded output
}

// ResponsePipe creates a limited buffer to reduce the number of dynamic
// resizing in ResponseWriter. We make the assumption that buffer cap is at
// least > than the average input message len
type ResponsePipe struct {
	dest http.ResponseWriter
	buf  []byte
}

// NewResponsePipe creates a ResponsePipe of capacity bufcap
func NewResponsePipe(resp http.ResponseWriter, bufcap int) *ResponsePipe {
	return &ResponsePipe{
		dest: resp,
		buf:  make([]byte, 0, bufcap),
	}
}

// Write implements io.Writer interface
func (resp *ResponsePipe) Write(input []byte) (int, error) {
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
	} else {
		copy(resp.buf[start:], input)
	}
}
