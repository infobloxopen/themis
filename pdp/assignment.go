package pdp

// AttributeAssignmentExpression represents assignment of arbitrary expression
// result to an attribute.
type AttributeAssignmentExpression struct {
	a Attribute
	e Expression
}

// MakeAttributeAssignmentExpression creates attribute assignment expression.
func MakeAttributeAssignmentExpression(a Attribute, e Expression) AttributeAssignmentExpression {
	return AttributeAssignmentExpression{
		a: a,
		e: e}
}

// Serialize evaluates assignment expression and returns string representation
// of resulting attribute name, type and value or error if the evaluaction
// can't be done.
func (a AttributeAssignmentExpression) Serialize(ctx *Context) (string, string, string, error) {
	ID := a.a.id
	typeName := BuiltinTypeKeys[a.a.t]

	v, err := a.e.Calculate(ctx)
	if err != nil {
		return ID, typeName, "", bindErrorf(err, "assignment to %q", ID)
	}

	t := v.GetResultType()
	if a.a.t != t {
		return ID, typeName, "", bindErrorf(newAssignmentTypeMismatch(a.a, t), "assignment to %q", ID)
	}

	s, err := v.Serialize()
	if err != nil {
		return ID, typeName, "", bindErrorf(err, "assignment to %q", ID)
	}

	return ID, typeName, s, nil
}
