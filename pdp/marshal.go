package pdp

import (
	"encoding/json"
	"io"
)

// StorageMarshal interface defines functions
// to capturing storage state information
type StorageMarshal interface {
	MarshalWithDepth(out io.Writer, depth int) error
	PathMarshal(ID string) (func(io.Writer) error, bool)
}

// PolicySet/Policy/Rule representation for marshaling
type storageNodeFmt struct {
	Ord int    `json:"ord"`
	ID  string `json:"id"`
}

func marshalHeader(v interface{}, out io.Writer) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	n := len(b)
	if n < 1 || b[n-1] != '}' {
		return newInvalidHeaderError(v)
	}
	_, err = out.Write(b[:n-1])
	return err
}
