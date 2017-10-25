package pdp_service

import (
	"github.com/valyala/gorpc"
)

// Effect field values
const (
	DENY            = 0
	PERMIT          = 1
	NOTAPPLICABLE   = 2
	INDETERMINATE   = 3
	INDETERMINATED  = 4
	INDETERMINATEP  = 5
	INDETERMINATEDP = 6
)

func EffectName(effect byte) string {
	switch effect {
	case 0:
		return "DENY"
	case 1:
		return "PERMIT"
	case 2:
		return "NOTAPPLICABLE"
	case 3:
		return "INDETERMINATE"
	case 4:
		return "INDETERMINATED"
	case 5:
		return "INDETERMINATEP"
	case 6:
		return "INDETERMINATEDP"
	}
	return "INVALID EFFECT"
}

type Attribute [3]string

func (a *Attribute) Id() string    { return a[0] }
func (a *Attribute) Type() string  { return a[1] }
func (a *Attribute) Value() string { return a[2] }

type Request struct {
	Attributes []*Attribute
}

type Response struct {
	Effect      byte
	Reason      string
	Obligations []*Attribute
}

func init() {
	gorpc.RegisterType(&Request{})
	gorpc.RegisterType(&Response{})
}
