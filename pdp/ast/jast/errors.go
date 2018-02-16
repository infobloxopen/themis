package jast

/* AUTOMATICALLY GENERATED FROM errors.yaml - DO NOT EDIT */

import (
	"encoding/json"
	"fmt"
	"github.com/infobloxopen/themis/pdp"
	"strings"
)

const (
	externalErrorID                     = 0
	attributeTypeErrorID                = 1
	policyAmbiguityErrorID              = 2
	policyMissingKeyErrorID             = 3
	unknownRCAErrorID                   = 4
	missingRCAErrorID                   = 5
	parseCAErrorID                      = 6
	invalidRCAErrorID                   = 7
	missingDefaultRuleRCAErrorID        = 8
	missingErrorRuleRCAErrorID          = 9
	notImplementedRCAErrorID            = 10
	unknownPCAErrorID                   = 11
	missingPCAErrorID                   = 12
	invalidPCAErrorID                   = 13
	missingDefaultPolicyPCAErrorID      = 14
	missingErrorPolicyPCAErrorID        = 15
	notImplementedPCAErrorID            = 16
	mapperArgumentTypeErrorID           = 17
	conditionTypeErrorID                = 18
	unknownEffectErrorID                = 19
	unknownMatchFunctionErrorID         = 20
	matchFunctionCastErrorID            = 21
	matchFunctionArgsNumberErrorID      = 22
	invalidMatchFunctionArgErrorID      = 23
	matchFunctionBothValuesErrorID      = 24
	matchFunctionBothAttrsErrorID       = 25
	unknownFunctionErrorID              = 26
	functionCastErrorID                 = 27
	unknownAttributeErrorID             = 28
	missingAttributeErrorID             = 29
	unknownMapperCAOrderID              = 30
	unknownTypeErrorID                  = 31
	invalidTypeErrorID                  = 32
	missingContentErrorID               = 33
	notImplementedValueTypeErrorID      = 34
	invalidAddressErrorID               = 35
	integerOverflowErrorID              = 36
	invalidNetworkErrorID               = 37
	invalidDomainErrorID                = 38
	selectorURIErrorID                  = 39
	selectorLocationErrorID             = 40
	unsupportedSelectorSchemeErrorID    = 41
	entityAmbiguityErrorID              = 42
	entityMissingKeyErrorID             = 43
	unknownPolicyUpdateOperationErrorID = 44
	missingContentTypeErrorID           = 45
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

type attributeTypeError struct {
	errorLink
	t string
}

func newAttributeTypeError(t string) *attributeTypeError {
	return &attributeTypeError{
		errorLink: errorLink{id: attributeTypeErrorID},
		t:         t}
}

func (e *attributeTypeError) Error() string {
	return e.errorf("Expected attribute data type but got \"%s\"", e.t)
}

type policyAmbiguityError struct {
	errorLink
}

func newPolicyAmbiguityError() *policyAmbiguityError {
	return &policyAmbiguityError{
		errorLink: errorLink{id: policyAmbiguityErrorID}}
}

func (e *policyAmbiguityError) Error() string {
	return e.errorf("Expected rules (for policy) or policies (for policy set) but got both")
}

type policyMissingKeyError struct {
	errorLink
}

func newPolicyMissingKeyError() *policyMissingKeyError {
	return &policyMissingKeyError{
		errorLink: errorLink{id: policyMissingKeyErrorID}}
}

func (e *policyMissingKeyError) Error() string {
	return e.errorf("Expected rules (for policy) or policies (for policy set) but got nothing")
}

type unknownRCAError struct {
	errorLink
	alg string
}

func newUnknownRCAError(alg string) *unknownRCAError {
	return &unknownRCAError{
		errorLink: errorLink{id: unknownRCAErrorID},
		alg:       alg}
}

func (e *unknownRCAError) Error() string {
	return e.errorf("Unknown rule combinig algorithm \"%s\"", e.alg)
}

type missingRCAError struct {
	errorLink
}

func newMissingRCAError() *missingRCAError {
	return &missingRCAError{
		errorLink: errorLink{id: missingRCAErrorID}}
}

func (e *missingRCAError) Error() string {
	return e.errorf("Missing policy combinig algorithm")
}

type parseCAError struct {
	errorLink
	token json.Token
}

func newParseCAError(token json.Token) *parseCAError {
	return &parseCAError{
		errorLink: errorLink{id: parseCAErrorID},
		token:     token}
}

func (e *parseCAError) Error() string {
	return e.errorf("Expected string or { object delimiter for combinig algorithm but got %T (%#v)", e.token, e.token)
}

type invalidRCAError struct {
	errorLink
	v interface{}
}

func newInvalidRCAError(v interface{}) *invalidRCAError {
	return &invalidRCAError{
		errorLink: errorLink{id: invalidRCAErrorID},
		v:         v}
}

func (e *invalidRCAError) Error() string {
	return e.errorf("Expected string or *caParams as policy combinig algorithm but got %T", e.v)
}

type missingDefaultRuleRCAError struct {
	errorLink
	ID string
}

func newMissingDefaultRuleRCAError(ID string) *missingDefaultRuleRCAError {
	return &missingDefaultRuleRCAError{
		errorLink: errorLink{id: missingDefaultRuleRCAErrorID},
		ID:        ID}
}

func (e *missingDefaultRuleRCAError) Error() string {
	return e.errorf("No rule with ID %q to use as default rule", e.ID)
}

type missingErrorRuleRCAError struct {
	errorLink
	ID string
}

func newMissingErrorRuleRCAError(ID string) *missingErrorRuleRCAError {
	return &missingErrorRuleRCAError{
		errorLink: errorLink{id: missingErrorRuleRCAErrorID},
		ID:        ID}
}

func (e *missingErrorRuleRCAError) Error() string {
	return e.errorf("No rule with ID %q to use as on error rule", e.ID)
}

type notImplementedRCAError struct {
	errorLink
	ID string
}

func newNotImplementedRCAError(ID string) *notImplementedRCAError {
	return &notImplementedRCAError{
		errorLink: errorLink{id: notImplementedRCAErrorID},
		ID:        ID}
}

func (e *notImplementedRCAError) Error() string {
	return e.errorf("Parsing for %q rule combinig algorithm hasn't been implemented yet", e.ID)
}

type unknownPCAError struct {
	errorLink
	alg string
}

func newUnknownPCAError(alg string) *unknownPCAError {
	return &unknownPCAError{
		errorLink: errorLink{id: unknownPCAErrorID},
		alg:       alg}
}

func (e *unknownPCAError) Error() string {
	return e.errorf("Unknown policy combinig algorithm \"%s\"", e.alg)
}

type missingPCAError struct {
	errorLink
}

func newMissingPCAError() *missingPCAError {
	return &missingPCAError{
		errorLink: errorLink{id: missingPCAErrorID}}
}

func (e *missingPCAError) Error() string {
	return e.errorf("Missing policy combinig algorithm")
}

type invalidPCAError struct {
	errorLink
	v interface{}
}

func newInvalidPCAError(v interface{}) *invalidPCAError {
	return &invalidPCAError{
		errorLink: errorLink{id: invalidPCAErrorID},
		v:         v}
}

func (e *invalidPCAError) Error() string {
	return e.errorf("Expected string or *caParams as policy combinig algorithm but got %T", e.v)
}

type missingDefaultPolicyPCAError struct {
	errorLink
	ID string
}

func newMissingDefaultPolicyPCAError(ID string) *missingDefaultPolicyPCAError {
	return &missingDefaultPolicyPCAError{
		errorLink: errorLink{id: missingDefaultPolicyPCAErrorID},
		ID:        ID}
}

func (e *missingDefaultPolicyPCAError) Error() string {
	return e.errorf("No policy with ID %q to use as default policy", e.ID)
}

type missingErrorPolicyPCAError struct {
	errorLink
	ID string
}

func newMissingErrorPolicyPCAError(ID string) *missingErrorPolicyPCAError {
	return &missingErrorPolicyPCAError{
		errorLink: errorLink{id: missingErrorPolicyPCAErrorID},
		ID:        ID}
}

func (e *missingErrorPolicyPCAError) Error() string {
	return e.errorf("No policy with ID %q to use as on error policy", e.ID)
}

type notImplementedPCAError struct {
	errorLink
	ID string
}

func newNotImplementedPCAError(ID string) *notImplementedPCAError {
	return &notImplementedPCAError{
		errorLink: errorLink{id: notImplementedPCAErrorID},
		ID:        ID}
}

func (e *notImplementedPCAError) Error() string {
	return e.errorf("Parsing for %q policy combinig algorithm hasn't been implemented yet", e.ID)
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
	return e.errorf("Expected %s, %s or %s as argument but got %s", pdp.TypeNames[pdp.TypeString], pdp.TypeNames[pdp.TypeSetOfStrings], pdp.TypeNames[pdp.TypeListOfStrings], pdp.TypeNames[e.actual])
}

type conditionTypeError struct {
	errorLink
	t int
}

func newConditionTypeError(t int) *conditionTypeError {
	return &conditionTypeError{
		errorLink: errorLink{id: conditionTypeErrorID},
		t:         t}
}

func (e *conditionTypeError) Error() string {
	return e.errorf("Expected %q as condition expression result but got %q", pdp.TypeNames[pdp.TypeBoolean], pdp.TypeNames[e.t])
}

type unknownEffectError struct {
	errorLink
	e string
}

func newUnknownEffectError(e string) *unknownEffectError {
	return &unknownEffectError{
		errorLink: errorLink{id: unknownEffectErrorID},
		e:         e}
}

func (e *unknownEffectError) Error() string {
	return e.errorf("Unknown rule effect %q", e.e)
}

type unknownMatchFunctionError struct {
	errorLink
	ID string
}

func newUnknownMatchFunctionError(ID string) *unknownMatchFunctionError {
	return &unknownMatchFunctionError{
		errorLink: errorLink{id: unknownMatchFunctionErrorID},
		ID:        ID}
}

func (e *unknownMatchFunctionError) Error() string {
	return e.errorf("Unknown match function %q", e.ID)
}

type matchFunctionCastError struct {
	errorLink
	ID     string
	first  int
	second int
}

func newMatchFunctionCastError(ID string, first, second int) *matchFunctionCastError {
	return &matchFunctionCastError{
		errorLink: errorLink{id: matchFunctionCastErrorID},
		ID:        ID,
		first:     first,
		second:    second}
}

func (e *matchFunctionCastError) Error() string {
	return e.errorf("No function %s for arguments %s and %s", e.ID, pdp.TypeNames[e.first], pdp.TypeNames[e.second])
}

type matchFunctionArgsNumberError struct {
	errorLink
	n int
}

func newMatchFunctionArgsNumberError(n int) *matchFunctionArgsNumberError {
	return &matchFunctionArgsNumberError{
		errorLink: errorLink{id: matchFunctionArgsNumberErrorID},
		n:         n}
}

func (e *matchFunctionArgsNumberError) Error() string {
	return e.errorf("Expected two arguments got %d", e.n)
}

type invalidMatchFunctionArgError struct {
	errorLink
	expr pdp.Expression
}

func newInvalidMatchFunctionArgError(expr pdp.Expression) *invalidMatchFunctionArgError {
	return &invalidMatchFunctionArgError{
		errorLink: errorLink{id: invalidMatchFunctionArgErrorID},
		expr:      expr}
}

func (e *invalidMatchFunctionArgError) Error() string {
	return e.errorf("Expected one immediate value and one attribute got %T", e.expr)
}

type matchFunctionBothValuesError struct {
	errorLink
}

func newMatchFunctionBothValuesError() *matchFunctionBothValuesError {
	return &matchFunctionBothValuesError{
		errorLink: errorLink{id: matchFunctionBothValuesErrorID}}
}

func (e *matchFunctionBothValuesError) Error() string {
	return e.errorf("Expected one immediate value and one attribute got both immediate values")
}

type matchFunctionBothAttrsError struct {
	errorLink
}

func newMatchFunctionBothAttrsError() *matchFunctionBothAttrsError {
	return &matchFunctionBothAttrsError{
		errorLink: errorLink{id: matchFunctionBothAttrsErrorID}}
}

func (e *matchFunctionBothAttrsError) Error() string {
	return e.errorf("Expected one immediate value and one attribute got both immediate values")
}

type unknownFunctionError struct {
	errorLink
	ID string
}

func newUnknownFunctionError(ID string) *unknownFunctionError {
	return &unknownFunctionError{
		errorLink: errorLink{id: unknownFunctionErrorID},
		ID:        ID}
}

func (e *unknownFunctionError) Error() string {
	return e.errorf("Unknown function %q", e.ID)
}

type functionCastError struct {
	errorLink
	ID    string
	exprs []pdp.Expression
}

func newFunctionCastError(ID string, exprs []pdp.Expression) *functionCastError {
	return &functionCastError{
		errorLink: errorLink{id: functionCastErrorID},
		ID:        ID,
		exprs:     exprs}
}

func (e *functionCastError) Error() string {
	args := ""
	if len(e.exprs) > 1 {
		t := make([]string, len(e.exprs))
		for i, e := range e.exprs {
			t[i] = pdp.TypeNames[e.GetResultType()]
		}
		args = fmt.Sprintf("%d arguments of following types \"%s\"", len(e.exprs), strings.Join(t, "\", \""))
	} else if len(e.exprs) > 0 {
		args = fmt.Sprintf("argument of type \"%s\"", pdp.TypeNames[e.exprs[0].GetResultType()])
	} else {
		args = "no arguments"
	}

	return e.errorf("Can't find function %s which takes %s", e.ID, args)
}

type unknownAttributeError struct {
	errorLink
	ID string
}

func newUnknownAttributeError(ID string) *unknownAttributeError {
	return &unknownAttributeError{
		errorLink: errorLink{id: unknownAttributeErrorID},
		ID:        ID}
}

func (e *unknownAttributeError) Error() string {
	return e.errorf("Unknown attribute %q", e.ID)
}

type missingAttributeError struct {
	errorLink
	attr string
	obj  string
}

func newMissingAttributeError(attr, obj string) *missingAttributeError {
	return &missingAttributeError{
		errorLink: errorLink{id: missingAttributeErrorID},
		attr:      attr,
		obj:       obj}
}

func (e *missingAttributeError) Error() string {
	return e.errorf("Missing %q attribute %q", e.obj, e.attr)
}

type unknownMapperCAOrder struct {
	errorLink
	ord string
}

func newUnknownMapperCAOrder(ord string) *unknownMapperCAOrder {
	return &unknownMapperCAOrder{
		errorLink: errorLink{id: unknownMapperCAOrderID},
		ord:       ord}
}

func (e *unknownMapperCAOrder) Error() string {
	return e.errorf("Unknown ordering for mapper \"%s\"", e.ord)
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

type invalidTypeError struct {
	errorLink
	t int
}

func newInvalidTypeError(t int) *invalidTypeError {
	return &invalidTypeError{
		errorLink: errorLink{id: invalidTypeErrorID},
		t:         t}
}

func (e *invalidTypeError) Error() string {
	return e.errorf("Can't make value of %q type", pdp.TypeNames[e.t])
}

type missingContentError struct {
	errorLink
}

func newMissingContentError() *missingContentError {
	return &missingContentError{
		errorLink: errorLink{id: missingContentErrorID}}
}

func (e *missingContentError) Error() string {
	return e.errorf("Missing value content")
}

type notImplementedValueTypeError struct {
	errorLink
	t int
}

func newNotImplementedValueTypeError(t int) *notImplementedValueTypeError {
	return &notImplementedValueTypeError{
		errorLink: errorLink{id: notImplementedValueTypeErrorID},
		t:         t}
}

func (e *notImplementedValueTypeError) Error() string {
	return e.errorf("Parsing for type %s hasn't been implemented yet", pdp.TypeNames[e.t])
}

type invalidAddressError struct {
	errorLink
	s string
}

func newInvalidAddressError(s string) *invalidAddressError {
	return &invalidAddressError{
		errorLink: errorLink{id: invalidAddressErrorID},
		s:         s}
}

func (e *invalidAddressError) Error() string {
	return e.errorf("Expected value of address type but got %q", e.s)
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

type invalidNetworkError struct {
	errorLink
	s   string
	err error
}

func newInvalidNetworkError(s string, err error) *invalidNetworkError {
	return &invalidNetworkError{
		errorLink: errorLink{id: invalidNetworkErrorID},
		s:         s,
		err:       err}
}

func (e *invalidNetworkError) Error() string {
	return e.errorf("Expected value of network type but got %q (%v)", e.s, e.err)
}

type invalidDomainError struct {
	errorLink
	s   string
	err error
}

func newInvalidDomainError(s string, err error) *invalidDomainError {
	return &invalidDomainError{
		errorLink: errorLink{id: invalidDomainErrorID},
		s:         s,
		err:       err}
}

func (e *invalidDomainError) Error() string {
	return e.errorf("Expected value of domain type but got %q (%v)", e.s, e.err)
}

type selectorURIError struct {
	errorLink
	uri string
	err error
}

func newSelectorURIError(uri string, err error) *selectorURIError {
	return &selectorURIError{
		errorLink: errorLink{id: selectorURIErrorID},
		uri:       uri,
		err:       err}
}

func (e *selectorURIError) Error() string {
	return e.errorf("Expected seletor URI but got %q (%s)", e.uri, e.err)
}

type selectorLocationError struct {
	errorLink
	loc string
	uri string
}

func newSelectorLocationError(loc, uri string) *selectorLocationError {
	return &selectorLocationError{
		errorLink: errorLink{id: selectorLocationErrorID},
		loc:       loc,
		uri:       uri}
}

func (e *selectorLocationError) Error() string {
	return e.errorf("Expected selector location in form of <Content-ID>/<Item-ID> got %q (%s)", e.loc, e.uri)
}

type unsupportedSelectorSchemeError struct {
	errorLink
	scheme string
	uri    string
}

func newUnsupportedSelectorSchemeError(scheme, uri string) *unsupportedSelectorSchemeError {
	return &unsupportedSelectorSchemeError{
		errorLink: errorLink{id: unsupportedSelectorSchemeErrorID},
		scheme:    scheme,
		uri:       uri}
}

func (e *unsupportedSelectorSchemeError) Error() string {
	return e.errorf("Unsupported selector scheme %q (%s)", e.scheme, e.uri)
}

type entityAmbiguityError struct {
	errorLink
	fields []string
}

func newEntityAmbiguityError(fields []string) *entityAmbiguityError {
	return &entityAmbiguityError{
		errorLink: errorLink{id: entityAmbiguityErrorID},
		fields:    fields}
}

func (e *entityAmbiguityError) Error() string {
	return e.errorf("Expected rules (for policy), policies (for policy set) or effect (for rule) but got %s", strings.Join(e.fields, ", "))
}

type entityMissingKeyError struct {
	errorLink
}

func newEntityMissingKeyError() *entityMissingKeyError {
	return &entityMissingKeyError{
		errorLink: errorLink{id: entityMissingKeyErrorID}}
}

func (e *entityMissingKeyError) Error() string {
	return e.errorf("Expected rules (for policy), policies (for policy set) or effect (for rule) but got nothing")
}

type unknownPolicyUpdateOperationError struct {
	errorLink
	op string
}

func newUnknownPolicyUpdateOperationError(op string) *unknownPolicyUpdateOperationError {
	return &unknownPolicyUpdateOperationError{
		errorLink: errorLink{id: unknownPolicyUpdateOperationErrorID},
		op:        op}
}

func (e *unknownPolicyUpdateOperationError) Error() string {
	return e.errorf("Unknown policy update operation %q", e.op)
}

type missingContentTypeError struct {
	errorLink
}

func newMissingContentTypeError() *missingContentTypeError {
	return &missingContentTypeError{
		errorLink: errorLink{id: missingContentTypeErrorID}}
}

func (e *missingContentTypeError) Error() string {
	return e.errorf("Value 'type' attribute is missing or placed after 'content' attribute")
}
