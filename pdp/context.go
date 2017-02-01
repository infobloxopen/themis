package pdp

import "fmt"

const (
	EffectDeny   = iota
	EffectPermit = iota

	EffectNotApplicable = iota

	EffectIndeterminate   = iota
	EffectIndeterminateD  = iota
	EffectIndeterminateP  = iota
	EffectIndeterminateDP = iota
)

var EffectNames map[int]string = map[int]string{
	EffectDeny:            "Deny",
	EffectPermit:          "Permit",
	EffectNotApplicable:   "NotApplicable",
	EffectIndeterminate:   "Indeterminate",
	EffectIndeterminateD:  "Indeterminate{D}",
	EffectIndeterminateP:  "Indeterminate{P}",
	EffectIndeterminateDP: "Indeterminate{DP}"}

var EffectIDs map[string]int = map[string]int{
	"deny":   EffectDeny,
	"permit": EffectPermit}

type Context struct {
	Attributes map[string]map[int]AttributeValueType
}

func NewContext() Context {
	return Context{make(map[string]map[int]AttributeValueType)}
}

func (c *Context) StoreAttribute(ID string, dataType int, a AttributeValueType) {
	t, ok := c.Attributes[ID]
	if ok {
		t[dataType] = a
		return
	}

	c.Attributes[ID] = map[int]AttributeValueType{dataType: a}
}

func (c *Context) GetAttribute(attr AttributeType) (AttributeValueType, error) {
	t, ok := c.Attributes[attr.ID]
	if !ok {
		return AttributeValueType{}, fmt.Errorf("Missing attribute %s (%s)", attr.ID, DataTypeNames[attr.DataType])
	}

	v, ok := t[attr.DataType]
	if !ok {
		return AttributeValueType{}, fmt.Errorf("Missing attribute %s (%s)", attr.ID, DataTypeNames[attr.DataType])
	}

	return v, nil
}

func (c *Context) CalculateObligations(obligations []AttributeAssignmentExpressionType, ctx *Context) error {
	for _, obligation := range obligations {
		v, err := obligation.Expression.calculate(ctx)
		if err != nil {
			return err
		}

		c.StoreAttribute(obligation.Attribute.ID, obligation.Attribute.DataType, v)
	}

	return nil
}

type ResponseType struct {
	Effect      int
	Status      string
	Obligations []AttributeAssignmentExpressionType
}

type EvaluableType interface {
	getID() string
	Calculate(ctx *Context) ResponseType
}

func combineEffectAndStatus(err error, ID string, r ResponseType) ResponseType {
	s := fmt.Sprintf("Match (%s): %s", ID, err)
	if r.Status != "Ok" {
		s = fmt.Sprintf("%s (%s)", s, r.Status)
	}

	if r.Effect == EffectNotApplicable {
		return ResponseType{EffectNotApplicable, s, nil}
	}

	if r.Effect == EffectDeny || r.Effect == EffectIndeterminateD {
		return ResponseType{EffectIndeterminateD, s, nil}
	}

	if r.Effect == EffectPermit || r.Effect == EffectIndeterminateP {
		return ResponseType{EffectIndeterminateP, s, nil}
	}

	return ResponseType{EffectIndeterminateDP, s, nil}
}
