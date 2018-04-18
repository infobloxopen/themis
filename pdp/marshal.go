package pdp

import "io"

// StorageMarshal interface defines functions
// to capturing storage state information
type StorageMarshal interface {
	MarshalJSON(out io.Writer, depth int) error
}

// PolicySet/Policy/Rule representation for marshaling
type storageNodeFmt struct {
	Ord int    `json:"ord"`
	ID  string `json:"id"`
}
