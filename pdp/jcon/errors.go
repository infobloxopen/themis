package jcon

/* AUTOMATICALLY GENERATED FROM errors.yaml - DO NOT EDIT */

import (
	"encoding/json"
	"github.com/infobloxopen/themis/pdp"
	"strings"
)

const (
	externalErrorID                       = 0
	objectKeyErrorID                      = 1
	booleanCastErrorID                    = 2
	numberCastErrorID                     = 3
	integerOverflowErrorID                = 4
	stringCastErrorID                     = 5
	addressCastErrorID                    = 6
	networkCastErrorID                    = 7
	domainCastErrorID                     = 8
	addressNetworkCastErrorID             = 9
	unknownContentFieldErrorID            = 10
	unknownContentItemFieldErrorID        = 11
	unknownTypeErrorID                    = 12
	invalidContentItemTypeErrorID         = 13
	invalidContentKeyTypeErrorID          = 14
	unknownDataFormatErrorID              = 15
	duplicateContentItemFieldErrorID      = 16
	missingContentDataErrorID             = 17
	missingContentTypeErrorID             = 18
	invalidSequenceContentItemNodeErrorID = 19
	invalidMapContentItemNodeErrorID      = 20
	unknownCommadFieldErrorID             = 21
	duplicateCommandFieldErrorID          = 22
	missingCommandOpErrorID               = 23
	missingCommandPathErrorID             = 24
	missingCommandEntityErrorID           = 25
	unknownContentUpdateOperationErrorID  = 26
	arrayEndDelimiterErrorID              = 27
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

type objectKeyError struct {
	errorLink
	token json.Token
}

func newObjectKeyError(token json.Token) *objectKeyError {
	return &objectKeyError{
		errorLink: errorLink{id: objectKeyErrorID},
		token:     token}
}

func (e *objectKeyError) Error() string {
	return e.errorf("Expected string as JSON object key but got %T (%#v)", e.token, e.token)
}

type booleanCastError struct {
	errorLink
	token json.Token
	desc  string
}

func newBooleanCastError(token json.Token, desc string) *booleanCastError {
	return &booleanCastError{
		errorLink: errorLink{id: booleanCastErrorID},
		token:     token,
		desc:      desc}
}

func (e *booleanCastError) Error() string {
	return e.errorf("Expected boolean as %s but got %T (%#v)", e.desc, e.token, e.token)
}

type numberCastError struct {
	errorLink
	token json.Token
	desc  string
}

func newNumberCastError(token json.Token, desc string) *numberCastError {
	return &numberCastError{
		errorLink: errorLink{id: numberCastErrorID},
		token:     token,
		desc:      desc}
}

func (e *numberCastError) Error() string {
	return e.errorf("Expected number as %s but got %T (%#v)", e.desc, e.token, e.token)
}

type integerOverflowError struct {
	errorLink
	x float64
}

func newIntegerOverflowError(x float64) *integerOverflowError {
	return &integerOverflowError{
		errorLink: errorLink{id: integerOverflowErrorID},
		x:         x}
}

func (e *integerOverflowError) Error() string {
	return e.errorf("%f overflows integer", e.x)
}

type stringCastError struct {
	errorLink
	token json.Token
	desc  string
}

func newStringCastError(token json.Token, desc string) *stringCastError {
	return &stringCastError{
		errorLink: errorLink{id: stringCastErrorID},
		token:     token,
		desc:      desc}
}

func (e *stringCastError) Error() string {
	return e.errorf("Expected string as %s but got %T (%#v)", e.desc, e.token, e.token)
}

type addressCastError struct {
	errorLink
	s string
}

func newAddressCastError(s string) *addressCastError {
	return &addressCastError{
		errorLink: errorLink{id: addressCastErrorID},
		s:         s}
}

func (e *addressCastError) Error() string {
	return e.errorf("Can't treat %q as IP address", e.s)
}

type networkCastError struct {
	errorLink
	s   string
	err error
}

func newNetworkCastError(s string, err error) *networkCastError {
	return &networkCastError{
		errorLink: errorLink{id: networkCastErrorID},
		s:         s,
		err:       err}
}

func (e *networkCastError) Error() string {
	return e.errorf("Can't treat %q as IP network (%s)", e.s, e.err)
}

type domainCastError struct {
	errorLink
	s   string
	err error
}

func newDomainCastError(s string, err error) *domainCastError {
	return &domainCastError{
		errorLink: errorLink{id: domainCastErrorID},
		s:         s,
		err:       err}
}

func (e *domainCastError) Error() string {
	return e.errorf("Can't treat %q as domain name (%s)", e.s, e.err)
}

type addressNetworkCastError struct {
	errorLink
	s   string
	err error
}

func newAddressNetworkCastError(s string, err error) *addressNetworkCastError {
	return &addressNetworkCastError{
		errorLink: errorLink{id: addressNetworkCastErrorID},
		s:         s,
		err:       err}
}

func (e *addressNetworkCastError) Error() string {
	return e.errorf("Can't treat %q as IP address or network (%s)", e.s, e.err)
}

type unknownContentFieldError struct {
	errorLink
	id string
}

func newUnknownContentFieldError(id string) *unknownContentFieldError {
	return &unknownContentFieldError{
		errorLink: errorLink{id: unknownContentFieldErrorID},
		id:        id}
}

func (e *unknownContentFieldError) Error() string {
	return e.errorf("Unknown content field %q (expected id or items)", e.id)
}

type unknownContentItemFieldError struct {
	errorLink
	id string
}

func newUnknownContentItemFieldError(id string) *unknownContentItemFieldError {
	return &unknownContentItemFieldError{
		errorLink: errorLink{id: unknownContentItemFieldErrorID},
		id:        id}
}

func (e *unknownContentItemFieldError) Error() string {
	return e.errorf("Unknown content item field %q (expected keys, type or data)", e.id)
}

type unknownTypeError struct {
	errorLink
	t string
}

func newUnknownTypeError(t string) *unknownTypeError {
	return &unknownTypeError{
		errorLink: errorLink{id: unknownTypeErrorID},
		t:         t}
}

func (e *unknownTypeError) Error() string {
	return e.errorf("Unknown value type %q", e.t)
}

type invalidContentItemTypeError struct {
	errorLink
	t int
}

func newInvalidContentItemTypeError(t int) *invalidContentItemTypeError {
	return &invalidContentItemTypeError{
		errorLink: errorLink{id: invalidContentItemTypeErrorID},
		t:         t}
}

func (e *invalidContentItemTypeError) Error() string {
	return e.errorf("Can't set result type to %q type", pdp.TypeNames[e.t])
}

type invalidContentKeyTypeError struct {
	errorLink
	t        int
	expected map[int]bool
}

func newInvalidContentKeyTypeError(t int, expected map[int]bool) *invalidContentKeyTypeError {
	return &invalidContentKeyTypeError{
		errorLink: errorLink{id: invalidContentKeyTypeErrorID},
		t:         t,
		expected:  expected}
}

func (e *invalidContentKeyTypeError) Error() string {
	names := make([]string, len(e.expected))
	i := 0
	for t := range e.expected {
		names[i] = pdp.TypeNames[t]
		i++
	}
	s := strings.Join(names, ", ")

	return e.errorf("Can't use %q type as a key in content item (expected %s)", pdp.TypeNames[e.t], s)
}

type unknownDataFormatError struct {
	errorLink
}

func newUnknownDataFormatError() *unknownDataFormatError {
	return &unknownDataFormatError{
		errorLink: errorLink{id: unknownDataFormatErrorID}}
}

func (e *unknownDataFormatError) Error() string {
	return e.errorf("Can't parse data without keys and result type information")
}

type duplicateContentItemFieldError struct {
	errorLink
	field string
}

func newDuplicateContentItemFieldError(field string) *duplicateContentItemFieldError {
	return &duplicateContentItemFieldError{
		errorLink: errorLink{id: duplicateContentItemFieldErrorID},
		field:     field}
}

func (e *duplicateContentItemFieldError) Error() string {
	return e.errorf("Duplicate content field %s", e.field)
}

type missingContentDataError struct {
	errorLink
}

func newMissingContentDataError() *missingContentDataError {
	return &missingContentDataError{
		errorLink: errorLink{id: missingContentDataErrorID}}
}

func (e *missingContentDataError) Error() string {
	return e.errorf("Missing data")
}

type missingContentTypeError struct {
	errorLink
}

func newMissingContentTypeError() *missingContentTypeError {
	return &missingContentTypeError{
		errorLink: errorLink{id: missingContentTypeErrorID}}
}

func (e *missingContentTypeError) Error() string {
	return e.errorf("Missing result type")
}

type invalidSequenceContentItemNodeError struct {
	errorLink
	node interface{}
	desc string
}

func newInvalidSequenceContentItemNodeError(node interface{}, desc string) *invalidSequenceContentItemNodeError {
	return &invalidSequenceContentItemNodeError{
		errorLink: errorLink{id: invalidSequenceContentItemNodeErrorID},
		node:      node,
		desc:      desc}
}

func (e *invalidSequenceContentItemNodeError) Error() string {
	return e.errorf("Expected array or object for %s but got %T", e.desc, e.node)
}

type invalidMapContentItemNodeError struct {
	errorLink
	node interface{}
	desc string
}

func newInvalidMapContentItemNodeError(node interface{}, desc string) *invalidMapContentItemNodeError {
	return &invalidMapContentItemNodeError{
		errorLink: errorLink{id: invalidMapContentItemNodeErrorID},
		node:      node,
		desc:      desc}
}

func (e *invalidMapContentItemNodeError) Error() string {
	return e.errorf("Expected object for %s but got %T", e.desc, e.node)
}

type unknownCommadFieldError struct {
	errorLink
	cmd string
}

func newUnknownCommadFieldError(cmd string) *unknownCommadFieldError {
	return &unknownCommadFieldError{
		errorLink: errorLink{id: unknownCommadFieldErrorID},
		cmd:       cmd}
}

func (e *unknownCommadFieldError) Error() string {
	return e.errorf("Unknown field %s", e.cmd)
}

type duplicateCommandFieldError struct {
	errorLink
	field string
}

func newDuplicateCommandFieldError(field string) *duplicateCommandFieldError {
	return &duplicateCommandFieldError{
		errorLink: errorLink{id: duplicateCommandFieldErrorID},
		field:     field}
}

func (e *duplicateCommandFieldError) Error() string {
	return e.errorf("Duplicate field %s", e.field)
}

type missingCommandOpError struct {
	errorLink
}

func newMissingCommandOpError() *missingCommandOpError {
	return &missingCommandOpError{
		errorLink: errorLink{id: missingCommandOpErrorID}}
}

func (e *missingCommandOpError) Error() string {
	return e.errorf("Missing operation")
}

type missingCommandPathError struct {
	errorLink
}

func newMissingCommandPathError() *missingCommandPathError {
	return &missingCommandPathError{
		errorLink: errorLink{id: missingCommandPathErrorID}}
}

func (e *missingCommandPathError) Error() string {
	return e.errorf("Missing path")
}

type missingCommandEntityError struct {
	errorLink
}

func newMissingCommandEntityError() *missingCommandEntityError {
	return &missingCommandEntityError{
		errorLink: errorLink{id: missingCommandEntityErrorID}}
}

func (e *missingCommandEntityError) Error() string {
	return e.errorf("Missing entity")
}

type unknownContentUpdateOperationError struct {
	errorLink
	op string
}

func newUnknownContentUpdateOperationError(op string) *unknownContentUpdateOperationError {
	return &unknownContentUpdateOperationError{
		errorLink: errorLink{id: unknownContentUpdateOperationErrorID},
		op:        op}
}

func (e *unknownContentUpdateOperationError) Error() string {
	return e.errorf("Unknown content update operation %q", e.op)
}

type arrayEndDelimiterError struct {
	errorLink
	actual   json.Delim
	expected string
	desc     string
}

func newArrayEndDelimiterError(actual json.Delim, expected, desc string) *arrayEndDelimiterError {
	return &arrayEndDelimiterError{
		errorLink: errorLink{id: arrayEndDelimiterErrorID},
		actual:    actual,
		expected:  expected,
		desc:      desc}
}

func (e *arrayEndDelimiterError) Error() string {
	return e.errorf("Expected %s JSON array end %q but got delimiter %q", e.desc, e.expected, e.actual)
}
