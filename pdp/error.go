package pdp

import (
	"fmt"
	"strings"
)

const errorSourcePathSeparator = ">"

const (
	externalErrorID = iota
	multiErrorID
	missingAttributeErrorID
	missingValueErrorID
	attributeValueTypeErrorID
	mapperArgumentTypeErrorID
	missingContentErrorID
	missingContentItemErrorID
	finalContentSubitemErrorID
	mapContentSubitemErrorID
)

type boundError interface {
	error
	bind(src string)
}

func bindError(err error, src string) boundError {
	b, ok := err.(boundError)
	if ok {
		b.bind(src)
		return b
	}

	return newExternalError(err, src)
}

type errorLink struct {
	id   int
	path []string
}

func (e *errorLink) errorf(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("#%02x (%s): %s", e.id, strings.Join(e.path, errorSourcePathSeparator), msg)
}

func (e *errorLink) bind(src string) {
	e.path = append([]string{src}, e.path...)
}

type externalError struct {
	errorLink
	err error
}

func newExternalError(err error, src string) *externalError {
	return &externalError{
		errorLink: errorLink{
			id:   externalErrorID,
			path: []string{src}},
		err: err}
}

func (e *externalError) Error() string {
	return e.errorf("%s", e.err)
}

type multiError struct {
	errorLink
	errs []error
}

func mewMultiError(errs []error, src string) *multiError {
	return &multiError{
		errorLink: errorLink{
			id:   multiErrorID,
			path: []string{src}},
		errs: errs}
}

func newNoSrcMultiError(errs ...error) *multiError {
	return &multiError{
		errorLink: errorLink{id: multiErrorID},
		errs:      errs}
}

func (e *multiError) Error() string {
	msgs := make([]string, len(e.errs))
	for i, err := range e.errs {
		msgs[i] = err.Error()
	}

	return e.errorf("multiple errors: %s", strings.Join(msgs, "\", \""))
}

type missingAttributeError struct {
	errorLink
}

func newMissingAttributeError(src string) error {
	return &missingAttributeError{
		errorLink: errorLink{
			id:   missingAttributeErrorID,
			path: []string{src}}}
}

func (e *missingAttributeError) Error() string {
	return e.errorf("Missing attribute")
}

type attributeValueTypeError struct {
	errorLink
	t int
	e int
}

func newAttributeValueTypeError(e, t int, src string) error {
	return &attributeValueTypeError{
		errorLink: errorLink{
			id:   attributeValueTypeErrorID,
			path: []string{src}},
		t: t,
		e: e}
}

func (e *attributeValueTypeError) Error() string {
	return e.errorf("Expected %s value but got %s", typeNames[e.e], typeNames[e.t])
}

type missingValueError struct {
	errorLink
}

func (e *missingValueError) Error() string {
	return e.errorf("Missing value")
}

type mapperArgumentTypeError struct {
	errorLink
	t int
}

func newMapperArgumentTypeError(t int) error {
	return &mapperArgumentTypeError{
		errorLink: errorLink{id: mapperArgumentTypeErrorID},
		t:         t}
}

func (e *mapperArgumentTypeError) Error() string {
	return fmt.Sprintf("Expected %s, %s or %s as argument but got %s",
		typeNames[typeString], typeNames[typeSetOfStrings], typeNames[typeListOfStrings],
		typeNames[e.t])
}

type missingContentError struct {
	errorLink
}

func newMissingContentError(src string) error {
	return &missingContentError{
		errorLink: errorLink{
			id:   missingContentErrorID,
			path: []string{src}}}
}

func (e *missingContentError) Error() string {
	return e.errorf("Missing content")
}

type missingContentItemError struct {
	errorLink
}

func newMissingContentItemError(src string) error {
	return &missingContentItemError{
		errorLink: errorLink{
			id:   missingContentItemErrorID,
			path: []string{src}}}
}

func (e *missingContentItemError) Error() string {
	return e.errorf("Missing content item")
}

type finalContentSubitemError struct {
	errorLink
}

func newFinalContentSubitemError(src string) error {
	return &finalContentSubitemError{
		errorLink: errorLink{
			id:   finalContentSubitemErrorID,
			path: []string{src}}}
}

func (e *finalContentSubitemError) Error() string {
	return e.errorf("Not a final value of the content")
}

type mapContentSubitemError struct {
	errorLink
}

func newMapContentSubitemError(src string) error {
	return &mapContentSubitemError{
		errorLink: errorLink{
			id:   mapContentSubitemErrorID,
			path: []string{src}}}
}

func (e *mapContentSubitemError) Error() string {
	return e.errorf("Not a map of the content")
}
