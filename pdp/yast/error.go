package yast

import (
	"fmt"
	"strings"

	"github.com/infobloxopen/themis/pdp"
)

const errorSourcePathSeparator = ">"

const (
	externalErrorID = iota
	rootKeysErrorID
	stringErrorID
	missingStringErrorID
	mapErrorID
	missingMapErrorID
	listErrorID
	missingListErrorID
	attributeTypeErrorID
	policyAmbiguityErrorID
	policyMissingKeyErrorID
	unknownRCAErrorID
	missingRCAErrorID
	invalidRCAErrorID
	missingMapRCAParamErrorID
	missingDefaultRuleRCAErrorID
	missingErrorRuleRCAErrorID
	notImplementedRCAErrorID
	unknownPCAErrorID
	missingPCAErrorID
	invalidPCAErrorID
	missingMapPCAParamErrorID
	missingDefaultPolicyPCAErrorID
	missingErrorPolicyPCAErrorID
	notImplementedPCAErrorID
	conditionTypeErrorID
	unknownEffectErrorID
	noSMPItemsErrorID
	tooManySMPItemsErrorID
	unknownMatchFunctionErrorID
	matchFunctionCastErrorID
	matchFunctionArgsNumberErrorID
	invalidMatchFunctionArgErrorID
	matchFunctionBothValuesErrorID
	matchFunctionBothAttrsErrorID
	unknownFunctionErrorID
	functionCastErrorID
	unknownAttributeErrorID
	unknownTypeErrorID
	invalidTypeErrorID
	missingContentErrorID
	notImplementedValueTypeErrorID
	invalidAddressErrorID
	invalidNetworkErrorID
	invalidDomainErrorID
	selectorURIErrorID
	selectorLocationErrorID
	unsupportedSelectorSchemeErrorID
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
	if len(e.path) > 0 {
		return fmt.Sprintf("#%02x (%s): %s", e.id, strings.Join(e.path, errorSourcePathSeparator), msg)
	}

	return fmt.Sprintf("#%02x: %s", e.id, msg)
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

type rootKeysError struct {
	errorLink
	keys []string
}

func newRootKeysError(m map[interface{}]interface{}) *rootKeysError {
	keys := make([]string, len(m))
	i := 0
	for key := range m {
		keys[i] = fmt.Sprintf("%v", key)
		i++
	}

	return &rootKeysError{
		errorLink: errorLink{id: rootKeysErrorID},
		keys:      keys}
}

func (e *rootKeysError) Error() string {
	return e.errorf("Expected attribute definitions and policies but got: %s", strings.Join(e.keys, ", "))
}

type stringError struct {
	errorLink
	t string
	e string
}

func newStringError(v interface{}, desc string) *stringError {
	return &stringError{
		errorLink: errorLink{id: stringErrorID},
		t:         fmt.Sprintf("%T", v),
		e:         desc}
}

func (e *stringError) Error() string {
	return e.errorf("Expected %s but got %s", e.e, e.t)
}

type missingStringError struct {
	errorLink
	s string
}

func newMissingStringError(desc string) *missingStringError {
	return &missingStringError{
		errorLink: errorLink{id: missingStringErrorID},
		s:         desc}
}

func (e *missingStringError) Error() string {
	return e.errorf("Missing %s", e.s)
}

type mapError struct {
	errorLink
	t string
	e string
}

func newMapError(v interface{}, desc string) *mapError {
	return &mapError{
		errorLink: errorLink{id: mapErrorID},
		t:         fmt.Sprintf("%T", v),
		e:         desc}
}

func (e *mapError) Error() string {
	return e.errorf("Expected %s but got %s", e.e, e.t)
}

type missingMapError struct {
	errorLink
	s string
}

func newMissingMapError(desc string) *missingMapError {
	return &missingMapError{
		errorLink: errorLink{id: missingMapErrorID},
		s:         desc}
}

func (e *missingMapError) Error() string {
	return e.errorf("Missing %s", e.s)
}

type listError struct {
	errorLink
	t string
	e string
}

func newListError(v interface{}, desc string) *listError {
	return &listError{
		errorLink: errorLink{id: listErrorID},
		t:         fmt.Sprintf("%T", v),
		e:         desc}
}

func (e *listError) Error() string {
	return e.errorf("Expected %s but got %s", e.e, e.t)
}

type missingListError struct {
	errorLink
	s string
}

func newMissingListError(desc string) *missingListError {
	return &missingListError{
		errorLink: errorLink{id: missingListErrorID},
		s:         desc}
}

func (e *missingListError) Error() string {
	return e.errorf("Missing %s", e.s)
}

type attributeTypeError struct {
	errorLink
	t string
}

func newAttributeTypeError(t string, src string) *attributeTypeError {
	return &attributeTypeError{
		errorLink: errorLink{
			id:   attributeTypeErrorID,
			path: []string{src}},
		t: t}
}

func (e *attributeTypeError) Error() string {
	return e.errorf("Expected attribute data type but got \"%s\"", e.t)
}

type policyAmbiguityError struct {
	errorLink
}

func newPolicyAmbiguityError(src string) *policyAmbiguityError {
	return &policyAmbiguityError{
		errorLink: errorLink{
			id:   policyAmbiguityErrorID,
			path: []string{src}}}
}

func (e *policyAmbiguityError) Error() string {
	return e.errorf("Expected rules (for policy) or policies (for policy set) but got both")
}

type policyMissingKeyError struct {
	errorLink
}

func newPolicyMissingKeyError(src string) *policyMissingKeyError {
	return &policyMissingKeyError{
		errorLink: errorLink{
			id:   policyMissingKeyErrorID,
			path: []string{src}}}
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
	return &missingRCAError{errorLink: errorLink{id: missingRCAErrorID}}
}

func (e *missingRCAError) Error() string {
	return e.errorf("Missing policy combinig algorithm")
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
	return e.errorf("Expected string or map as policy combinig algorithm but got %T", e.v)
}

type missingMapRCAParamError struct {
	errorLink
}

func newMissingMapRCAParamError() *missingMapRCAParamError {
	return &missingMapRCAParamError{errorLink: errorLink{id: missingMapRCAParamErrorID}}
}

func (e *missingMapRCAParamError) Error() string {
	return e.errorf("Missing map parameter")
}

type missingDefaultRuleRCAError struct {
	errorLink
	n string
}

func newMissingDefaultRuleRCAError(n string) *missingDefaultRuleRCAError {
	return &missingDefaultRuleRCAError{
		errorLink: errorLink{id: missingDefaultRuleRCAErrorID},
		n:         n}
}

func (e *missingDefaultRuleRCAError) Error() string {
	return e.errorf("No rule with ID %q to use as default rule", e.n)
}

type missingErrorRuleRCAError struct {
	errorLink
	n string
}

func newMissingErrorRuleRCAError(n string) *missingErrorRuleRCAError {
	return &missingErrorRuleRCAError{
		errorLink: errorLink{id: missingErrorRuleRCAErrorID},
		n:         n}
}

func (e *missingErrorRuleRCAError) Error() string {
	return e.errorf("No rule with ID %q to use as on error rule", e.n)
}

type notImplementedRCAError struct {
	errorLink
	n string
}

func newNotImplementedRCAError(n string) *notImplementedRCAError {
	return &notImplementedRCAError{
		errorLink: errorLink{id: notImplementedRCAErrorID},
		n:         n}
}

func (e *notImplementedRCAError) Error() string {
	return e.errorf("Parsing for %q rule combinig algorithm hasn't been implemented yet", e.n)
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
	return &missingPCAError{errorLink: errorLink{id: missingPCAErrorID}}
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
	return e.errorf("Expected string or map as policy combinig algorithm but got %T", e.v)
}

type missingMapPCAParamError struct {
	errorLink
}

func newMissingMapPCAParamError() *missingMapPCAParamError {
	return &missingMapPCAParamError{errorLink: errorLink{id: missingMapPCAParamErrorID}}
}

func (e *missingMapPCAParamError) Error() string {
	return e.errorf("Missing map parameter")
}

type missingDefaultPolicyPCAError struct {
	errorLink
	n string
}

func newMissingDefaultPolicyPCAError(n string) *missingDefaultPolicyPCAError {
	return &missingDefaultPolicyPCAError{
		errorLink: errorLink{id: missingDefaultPolicyPCAErrorID},
		n:         n}
}

func (e *missingDefaultPolicyPCAError) Error() string {
	return e.errorf("No policy with ID %q to use as default policy", e.n)
}

type missingErrorPolicyPCAError struct {
	errorLink
	n string
}

func newMissingErrorPolicyPCAError(n string) *missingErrorPolicyPCAError {
	return &missingErrorPolicyPCAError{
		errorLink: errorLink{id: missingErrorPolicyPCAErrorID},
		n:         n}
}

func (e *missingErrorPolicyPCAError) Error() string {
	return e.errorf("No policy with ID %q to use as on error policy", e.n)
}

type notImplementedPCAError struct {
	errorLink
	n string
}

func newNotImplementedPCAError(n string) *notImplementedPCAError {
	return &notImplementedPCAError{
		errorLink: errorLink{id: notImplementedPCAErrorID},
		n:         n}
}

func (e *notImplementedPCAError) Error() string {
	return e.errorf("Parsing for %q policy combinig algorithm hasn't been implemented yet", e.n)
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
	return e.errorf("Expected %q as condition expression result but got %q",
		pdp.TypeNames[pdp.TypeBoolean], pdp.TypeNames[e.t])
}

type unknownEffectError struct {
	errorLink
	e string
}

func newUnknownEffectError(e string, src string) *unknownEffectError {
	return &unknownEffectError{
		errorLink: errorLink{
			id:   unknownEffectErrorID,
			path: []string{src}},
		e: e}
}

func (e *unknownEffectError) Error() string {
	return e.errorf("Unknown rule effect \"%s\"", e.e)
}

type noSMPItemsError struct {
	errorLink
	e string
	n int
}

func newNoSMPItemsError(e string, n int) *noSMPItemsError {
	return &noSMPItemsError{
		errorLink: errorLink{id: noSMPItemsErrorID},
		e:         e,
		n:         n}
}

func (e *noSMPItemsError) Error() string {
	return e.errorf("Expected at least one entry in %s got %d", e.e, e.n)
}

type tooManySMPItemsError struct {
	errorLink
	e string
	n int
}

func newTooManySMPItemsError(e string, n int) *tooManySMPItemsError {
	return &tooManySMPItemsError{
		errorLink: errorLink{id: tooManySMPItemsErrorID},
		e:         e,
		n:         n}
}

func (e *tooManySMPItemsError) Error() string {
	return e.errorf("Expected only one entry in %s got %d", e.e, e.n)
}

type unknownMatchFunctionError struct {
	errorLink
	n string
}

func newUnknownMatchFunctionError(n string) *unknownMatchFunctionError {
	return &unknownMatchFunctionError{
		errorLink: errorLink{id: unknownMatchFunctionErrorID},
		n:         n}
}

func (e *unknownMatchFunctionError) Error() string {
	return e.errorf("Unknown match function %q", e.n)
}

type matchFunctionCastError struct {
	errorLink
	n  string
	a1 string
	a2 string
}

func newMatchFunctionCastError(n, a1, a2 string) *matchFunctionCastError {
	return &matchFunctionCastError{
		errorLink: errorLink{
			id:   matchFunctionCastErrorID,
			path: []string{n}},
		n:  n,
		a1: a1,
		a2: a2}
}

func (e *matchFunctionCastError) Error() string {
	return e.errorf("No function %s for arguments %s and %s", e.n, e.a1, e.a2)
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
	e pdp.Expression
}

func newInvalidMatchFunctionArgError(e pdp.Expression) *invalidMatchFunctionArgError {
	return &invalidMatchFunctionArgError{
		errorLink: errorLink{id: invalidMatchFunctionArgErrorID},
		e:         e}
}

func (e *invalidMatchFunctionArgError) Error() string {
	return e.errorf("Expected one immediate value and one attribute got %T", e.e)
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
	n string
}

func newUnknownFunctionError(n string) *unknownFunctionError {
	return &unknownFunctionError{
		errorLink: errorLink{id: unknownFunctionErrorID},
		n:         n}
}

func (e *unknownFunctionError) Error() string {
	return e.errorf("Unknown function %q", e.n)
}

type functionCastError struct {
	errorLink
	n string
	e []pdp.Expression
}

func newFunctionCastError(n string, e []pdp.Expression) *functionCastError {
	return &functionCastError{
		errorLink: errorLink{id: functionCastErrorID},
		n:         n,
		e:         e}
}

func (e *functionCastError) Error() string {
	if len(e.e) > 1 {
		t := make([]string, len(e.e))
		for i, e := range e.e {
			t[i] = pdp.TypeNames[e.GetResultType()]
		}

		return e.errorf("Can't find function %s which takes %d arguments of following types \"%s\"",
			e.n, len(e.e), strings.Join(t, "\", \""))
	}

	if len(e.e) > 0 {
		return e.errorf("Can't find function %s which takes argument of type \"%s\"",
			e.n, pdp.TypeNames[e.e[0].GetResultType()])
	}

	return e.errorf("Can't find function %s which takes no arguments", e.n)
}

type unknownAttributeError struct {
	errorLink
	n string
}

func newUnknownAttributeError(n string) *unknownAttributeError {
	return &unknownAttributeError{
		errorLink: errorLink{id: unknownAttributeErrorID},
		n:         n}
}

func (e *unknownAttributeError) Error() string {
	return e.errorf("Unknown attribute %q", e.n)
}

type unknownTypeError struct {
	errorLink
	s string
}

func newUnknownTypeError(s string) *unknownTypeError {
	return &unknownTypeError{
		errorLink: errorLink{id: unknownTypeErrorID},
		s:         s}
}

func (e *unknownTypeError) Error() string {
	return e.errorf("Unknown value type %q", e.s)
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
	return &missingContentError{errorLink: errorLink{id: missingContentErrorID}}
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
