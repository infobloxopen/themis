package pdp

import "io"

// StorageMarshal interface defines functions
// to capturing storage state information
type StorageMarshal interface {
	DepthMarshal(out io.Writer, depth int) error
	PathMarshal(ID string) (func(io.Writer) error, bool)
}
