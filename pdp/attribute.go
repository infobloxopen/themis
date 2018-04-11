package pdp

import "fmt"

// Attribute represents attribute definition which binds attribute name
// and type.
type Attribute struct {
	id string
	t  int
}

// MakeAttribute creates new attribute instance. It requires attribute name
// as "ID" argument and type as "t" argument. Value of "t" should be one of
// Type* constants.
func MakeAttribute(ID string, t int) Attribute {
	return Attribute{id: ID, t: t}
}

// GetType returns attribute type.
func (a Attribute) GetType() int {
	return a.t
}

func (a Attribute) describe() string {
	return fmt.Sprintf("attr(%s.%s)", a.id, BuiltinTypeNames[a.t])
}
