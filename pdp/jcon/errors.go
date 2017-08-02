package jcon

/* AUTOMATICALLY GENERATED FROM errors.yaml - DO NOT EDIT */

import (
	"encoding/json"
	"github.com/infobloxopen/themis/pdp"
	"strings"
)

const (
	externalErrorID = iota
	rootObjectStartTokenErrorID
	rootObjectStartDelimiterErrorID
	objectStartTokenErrorID
	objectStartDelimiterErrorID
	objectEndDelimiterErrorID
	objectTokenErrorID
	arrayStartTokenErrorID
	arrayStartDelimiterErrorID
	arrayEndDelimiterErrorID
	stringArrayTokenErrorID
	objectArrayStartTokenErrorID
	objectArrayStartDelimiterErrorID
	unexpectedDelimiterErrorID
	objectKeyErrorID
	missingEOFErrorID
	booleanCastErrorID
	stringCastErrorID
	addressCastErrorID
	networkCastErrorID
	domainCastErrorID
	addressNetworkCastErrorID
	unknownContentFieldErrorID
	unknownContentItemFieldErrorID
	unknownTypeErrorID
	invalidContentItemTypeErrorID
	invalidContentKeyTypeErrorID
	unknownDataFormatErrorID
	duplicateContentItemFieldErrorID
	missingContentDataErrorID
	missingContentTypeErrorID
	invalidSequenceContentItemNodeErrorID
	invalidMapContentItemNodeErrorID
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

type rootObjectStartTokenError struct {
	errorLink
	actual   json.Token
	expected string
}

func newRootObjectStartTokenError(actual json.Token, expected string) *rootObjectStartTokenError {
	return &rootObjectStartTokenError{
		errorLink: errorLink{id: rootObjectStartTokenErrorID},
		actual:    actual,
		expected:  expected}
}

func (e *rootObjectStartTokenError) Error() string {
	return e.errorf("Expected root JSON object start %q but got token %T (%#v)", e.expected, e.actual, e.actual)
}

type rootObjectStartDelimiterError struct {
	errorLink
	actual   json.Delim
	expected string
}

func newRootObjectStartDelimiterError(actual json.Delim, expected string) *rootObjectStartDelimiterError {
	return &rootObjectStartDelimiterError{
		errorLink: errorLink{id: rootObjectStartDelimiterErrorID},
		actual:    actual,
		expected:  expected}
}

func (e *rootObjectStartDelimiterError) Error() string {
	return e.errorf("Expected root JSON object start %q but got delimiter %q", e.expected, e.actual)
}

type objectStartTokenError struct {
	errorLink
	actual   json.Token
	expected string
	desc     string
}

func newObjectStartTokenError(actual json.Token, expected, desc string) *objectStartTokenError {
	return &objectStartTokenError{
		errorLink: errorLink{id: objectStartTokenErrorID},
		actual:    actual,
		expected:  expected,
		desc:      desc}
}

func (e *objectStartTokenError) Error() string {
	return e.errorf("Expected %s JSON object start %q but got token %T (%#v)", e.desc, e.expected, e.actual, e.actual)
}

type objectStartDelimiterError struct {
	errorLink
	actual   json.Delim
	expected string
	desc     string
}

func newObjectStartDelimiterError(actual json.Delim, expected, desc string) *objectStartDelimiterError {
	return &objectStartDelimiterError{
		errorLink: errorLink{id: objectStartDelimiterErrorID},
		actual:    actual,
		expected:  expected,
		desc:      desc}
}

func (e *objectStartDelimiterError) Error() string {
	return e.errorf("Expected %s JSON object start %q but got delimiter %q", e.desc, e.expected, e.actual)
}

type objectEndDelimiterError struct {
	errorLink
	actual   json.Delim
	expected string
	desc     string
}

func newObjectEndDelimiterError(actual json.Delim, expected, desc string) *objectEndDelimiterError {
	return &objectEndDelimiterError{
		errorLink: errorLink{id: objectEndDelimiterErrorID},
		actual:    actual,
		expected:  expected,
		desc:      desc}
}

func (e *objectEndDelimiterError) Error() string {
	return e.errorf("Expected %s JSON object end %q but got delimiter %q", e.desc, e.expected, e.actual)
}

type objectTokenError struct {
	errorLink
	actual   json.Token
	expected string
	desc     string
}

func newObjectTokenError(actual json.Token, expected, desc string) *objectTokenError {
	return &objectTokenError{
		errorLink: errorLink{id: objectTokenErrorID},
		actual:    actual,
		expected:  expected,
		desc:      desc}
}

func (e *objectTokenError) Error() string {
	return e.errorf("Expected %s JSON object string key or end %q but got token %T (%#v)", e.desc, e.expected, e.actual, e.actual)
}

type arrayStartTokenError struct {
	errorLink
	actual   json.Token
	expected string
	desc     string
}

func newArrayStartTokenError(actual json.Token, expected, desc string) *arrayStartTokenError {
	return &arrayStartTokenError{
		errorLink: errorLink{id: arrayStartTokenErrorID},
		actual:    actual,
		expected:  expected,
		desc:      desc}
}

func (e *arrayStartTokenError) Error() string {
	return e.errorf("Expected %s JSON array start %q but got token %T (%#v)", e.desc, e.expected, e.actual, e.actual)
}

type arrayStartDelimiterError struct {
	errorLink
	actual   json.Delim
	expected string
	desc     string
}

func newArrayStartDelimiterError(actual json.Delim, expected, desc string) *arrayStartDelimiterError {
	return &arrayStartDelimiterError{
		errorLink: errorLink{id: arrayStartDelimiterErrorID},
		actual:    actual,
		expected:  expected,
		desc:      desc}
}

func (e *arrayStartDelimiterError) Error() string {
	return e.errorf("Expected %s JSON array start %q but got delimiter %q", e.desc, e.expected, e.actual)
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

type stringArrayTokenError struct {
	errorLink
	actual   json.Token
	expected string
	desc     string
}

func newStringArrayTokenError(actual json.Token, expected, desc string) *stringArrayTokenError {
	return &stringArrayTokenError{
		errorLink: errorLink{id: stringArrayTokenErrorID},
		actual:    actual,
		expected:  expected,
		desc:      desc}
}

func (e *stringArrayTokenError) Error() string {
	return e.errorf("Expected %s JSON array string value or end %q but got token %T (%#v)", e.desc, e.expected, e.actual, e.actual)
}

type objectArrayStartTokenError struct {
	errorLink
	actual         json.Token
	firstExpected  string
	secondExpected string
	desc           string
}

func newObjectArrayStartTokenError(actual json.Token, firstExpected, secondExpected, desc string) *objectArrayStartTokenError {
	return &objectArrayStartTokenError{
		errorLink:      errorLink{id: objectArrayStartTokenErrorID},
		actual:         actual,
		firstExpected:  firstExpected,
		secondExpected: secondExpected,
		desc:           desc}
}

func (e *objectArrayStartTokenError) Error() string {
	return e.errorf("Expected %s JSON object or array start %q or %q but got token %T (%#v)", e.desc, e.firstExpected, e.secondExpected, e.actual, e.actual)
}

type objectArrayStartDelimiterError struct {
	errorLink
	actual         json.Delim
	firstExpected  string
	secondExpected string
	desc           string
}

func newObjectArrayStartDelimiterError(actual json.Delim, firstExpected, secondExpected, desc string) *objectArrayStartDelimiterError {
	return &objectArrayStartDelimiterError{
		errorLink:      errorLink{id: objectArrayStartDelimiterErrorID},
		actual:         actual,
		firstExpected:  firstExpected,
		secondExpected: secondExpected,
		desc:           desc}
}

func (e *objectArrayStartDelimiterError) Error() string {
	return e.errorf("Expected %s JSON object or array start %q or %q but got delimiter %q", e.desc, e.firstExpected, e.secondExpected, e.actual)
}

type unexpectedDelimiterError struct {
	errorLink
	delim string
	desc  string
}

func newUnexpectedDelimiterError(delim, desc string) *unexpectedDelimiterError {
	return &unexpectedDelimiterError{
		errorLink: errorLink{id: unexpectedDelimiterErrorID},
		delim:     delim,
		desc:      desc}
}

func (e *unexpectedDelimiterError) Error() string {
	return e.errorf("Unexpected delimiter %q for %s", e.delim, e.desc)
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

type missingEOFError struct {
	errorLink
	token json.Token
}

func newMissingEOFError(token json.Token) *missingEOFError {
	return &missingEOFError{
		errorLink: errorLink{id: missingEOFErrorID},
		token:     token}
}

func (e *missingEOFError) Error() string {
	return e.errorf("Expected expected EOF after root object end but got %T (%#v)", e.token, e.token)
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
