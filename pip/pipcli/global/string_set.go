package global

import "strings"

// StringSet implements flag.Value inteface to accept array of strings as
// command line argument.
type StringSet []string

// String returns human readable representation of the array.
func (s *StringSet) String() string {
	return strings.Join(*s, ", ")
}

// Set appends given value to the end of array.
func (s *StringSet) Set(v string) error {
	*s = append(*s, v)
	return nil
}
