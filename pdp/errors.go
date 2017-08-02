package pdp

/* AUTOMATICALLY GENERATED FROM errors.yaml - DO NOT EDIT */

import "strings"

const (
	externalErrorID = iota
	multiErrorID
	missingAttributeErrorID
	missingValueErrorID
	attributeValueTypeErrorID
	mapperArgumentTypeErrorID
	missingContentErrorID
	missingContentItemErrorID
	invalidContentItemTypeErrorID
	invalidSelectorPathErrorID
	mapContentSubitemErrorID
)

type externalError struct {
	errorLink
	err error
}

func newExternalError(err error) *externalError {
	return &externalError{
		errorLink: errorLink{id: externalErrorID},
		err:       err}
}

func (e *externalError) Error() string {
	return e.errorf("%s", e.err)
}

type multiError struct {
	errorLink
	errs []error
}

func newMultiError(errs []error) *multiError {
	return &multiError{
		errorLink: errorLink{id: multiErrorID},
		errs:      errs}
}

func (e *multiError) Error() string {
	msgs := make([]string, len(e.errs))
	for i, err := range e.errs {
		msgs[i] = err.Error()
	}
	msg := strings.Join(msgs, ", ")

	return e.errorf("multiple errors: %s", msg)
}

type missingAttributeError struct {
	errorLink
}

func newMissingAttributeError() *missingAttributeError {
	return &missingAttributeError{
		errorLink: errorLink{id: missingAttributeErrorID}}
}

func (e *missingAttributeError) Error() string {
	return e.errorf("Missing attribute")
}

type missingValueError struct {
	errorLink
}

func newMissingValueError() *missingValueError {
	return &missingValueError{
		errorLink: errorLink{id: missingValueErrorID}}
}

func (e *missingValueError) Error() string {
	return e.errorf("Missing value")
}

type attributeValueTypeError struct {
	errorLink
	expected int
	actual   int
}

func newAttributeValueTypeError(expected, actual int) *attributeValueTypeError {
	return &attributeValueTypeError{
		errorLink: errorLink{id: attributeValueTypeErrorID},
		expected:  expected,
		actual:    actual}
}

func (e *attributeValueTypeError) Error() string {
	return e.errorf("Expected %s value but got %s", TypeNames[e.expected], TypeNames[e.actual])
}

type mapperArgumentTypeError struct {
	errorLink
	actual int
}

func newMapperArgumentTypeError(actual int) *mapperArgumentTypeError {
	return &mapperArgumentTypeError{
		errorLink: errorLink{id: mapperArgumentTypeErrorID},
		actual:    actual}
}

func (e *mapperArgumentTypeError) Error() string {
	return e.errorf("Expected %s, %s or %s as argument but got %s", TypeNames[TypeString], TypeNames[TypeSetOfStrings], TypeNames[TypeListOfStrings], TypeNames[e.actual])
}

type missingContentError struct {
	errorLink
}

func newMissingContentError() *missingContentError {
	return &missingContentError{
		errorLink: errorLink{id: missingContentErrorID}}
}

func (e *missingContentError) Error() string {
	return e.errorf("Missing content")
}

type missingContentItemError struct {
	errorLink
}

func newMissingContentItemError() *missingContentItemError {
	return &missingContentItemError{
		errorLink: errorLink{id: missingContentItemErrorID}}
}

func (e *missingContentItemError) Error() string {
	return e.errorf("Missing content item")
}

type invalidContentItemTypeError struct {
	errorLink
	expected int
	actual   int
}

func newInvalidContentItemTypeError(expected, actual int) *invalidContentItemTypeError {
	return &invalidContentItemTypeError{
		errorLink: errorLink{id: invalidContentItemTypeErrorID},
		expected:  expected,
		actual:    actual}
}

func (e *invalidContentItemTypeError) Error() string {
	return e.errorf("Invalid conent item type. Expected %q but got %q", TypeNames[e.expected], TypeNames[e.actual])
}

type invalidSelectorPathError struct {
	errorLink
	expected []int
	actual   []Expression
}

func newInvalidSelectorPathError(expected []int, actual []Expression) *invalidSelectorPathError {
	return &invalidSelectorPathError{
		errorLink: errorLink{id: invalidSelectorPathErrorID},
		expected:  expected,
		actual:    actual}
}

func (e *invalidSelectorPathError) Error() string {
	expStrs := make([]string, len(e.expected))
	for i, t := range e.expected {
		expStrs[i] = TypeNames[t]
	}
	expected := strings.Join(expStrs, "/")

	actStrs := make([]string, len(e.actual))
	for i, e := range e.actual {
		actStrs[i] = TypeNames[e.GetResultType()]
	}
	actual := strings.Join(actStrs, "/")

	return e.errorf("Invalid selector path. Expected %s but got %s", expected, actual)
}

type mapContentSubitemError struct {
	errorLink
}

func newMapContentSubitemError() *mapContentSubitemError {
	return &mapContentSubitemError{
		errorLink: errorLink{id: mapContentSubitemErrorID}}
}

func (e *mapContentSubitemError) Error() string {
	return e.errorf("Not a map of the content")
}
