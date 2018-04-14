package pdp

/* AUTOMATICALLY GENERATED FROM errors.yaml - DO NOT EDIT */

import (
	"github.com/google/uuid"
	"strconv"
	"strings"
)

// Numeric identifiers of errors.
const (
	externalErrorID                           = 0
	multiErrorID                              = 1
	missingAttributeErrorID                   = 2
	missingValueErrorID                       = 3
	unknownTypeStringCastErrorID              = 4
	invalidTypeStringCastErrorID              = 5
	notImplementedStringCastErrorID           = 6
	invalidBooleanStringCastErrorID           = 7
	invalidIntegerStringCastErrorID           = 8
	invalidFloatStringCastErrorID             = 9
	invalidAddressStringCastErrorID           = 10
	invalidNetworkStringCastErrorID           = 11
	invalidAddressNetworkStringCastErrorID    = 12
	invalidDomainNameStringCastErrorID        = 13
	attributeValueTypeErrorID                 = 14
	attributeValueFlagsTypeErrorID            = 15
	attributeValueFlagsBitsErrorID            = 16
	duplicateAttributeValueErrorID            = 17
	unknownTypeSerializationErrorID           = 18
	invalidTypeSerializationErrorID           = 19
	assignmentTypeMismatchID                  = 20
	mapperArgumentTypeErrorID                 = 21
	UntaggedPolicyModificationErrorID         = 22
	MissingPolicyTagErrorID                   = 23
	PolicyTagsNotMatchErrorID                 = 24
	emptyPathModificationErrorID              = 25
	invalidRootPolicyItemTypeErrorID          = 26
	hiddenRootPolicyAppendErrorID             = 27
	invalidRootPolicyErrorID                  = 28
	hiddenPolicySetModificationErrorID        = 29
	invalidPolicySetItemTypeErrorID           = 30
	tooShortPathPolicySetModificationErrorID  = 31
	missingPolicySetChildErrorID              = 32
	hiddenPolicyAppendErrorID                 = 33
	policyTransactionTagsNotMatchErrorID      = 34
	failedPolicyTransactionErrorID            = 35
	unknownPolicyUpdateOperationErrorID       = 36
	hiddenPolicyModificationErrorID           = 37
	tooLongPathPolicyModificationErrorID      = 38
	tooShortPathPolicyModificationErrorID     = 39
	invalidPolicyItemTypeErrorID              = 40
	hiddenRuleAppendErrorID                   = 41
	missingPolicyChildErrorID                 = 42
	missingContentErrorID                     = 43
	invalidContentStorageItemID               = 44
	missingContentItemErrorID                 = 45
	invalidContentItemErrorID                 = 46
	invalidContentItemTypeErrorID             = 47
	invalidSelectorPathErrorID                = 48
	networkMapKeyValueTypeErrorID             = 49
	mapContentSubitemErrorID                  = 50
	invalidContentModificationErrorID         = 51
	missingPathContentModificationErrorID     = 52
	tooLongPathContentModificationErrorID     = 53
	invalidContentValueModificationErrorID    = 54
	UntaggedContentModificationErrorID        = 55
	MissingContentTagErrorID                  = 56
	ContentTagsNotMatchErrorID                = 57
	unknownContentUpdateOperationErrorID      = 58
	failedContentTransactionErrorID           = 59
	contentTransactionIDNotMatchErrorID       = 60
	contentTransactionTagsNotMatchErrorID     = 61
	tooShortRawPathContentModificationErrorID = 62
	tooLongRawPathContentModificationErrorID  = 63
	invalidContentUpdateDataErrorID           = 64
	invalidContentUpdateResultTypeErrorID     = 65
	invalidContentUpdateKeysErrorID           = 66
	unknownContentItemResultTypeErrorID       = 67
	invalidContentItemResultTypeErrorID       = 68
	invalidContentKeyTypeErrorID              = 69
	invalidContentStringMapErrorID            = 70
	invalidContentNetworkMapErrorID           = 71
	invalidContentDomainMapErrorID            = 72
	invalidContentValueErrorID                = 73
	invalidContentValueTypeErrorID            = 74
	integerDivideByZeroErrorID                = 75
	floatDivideByZeroErrorID                  = 76
	floatNanErrorID                           = 77
	floatInfErrorID                           = 78
	nilTypeErrorID                            = 79
	builtinCustomTypeErrorID                  = 80
	duplicateCustomTypeErrorID                = 81
	duplicatesBuiltinTypeErrorID              = 82
	duplicateFlagNameID                       = 83
	noTypedAttributeErrorID                   = 84
	undefinedAttributeTypeErrorID             = 85
	unknownAttributeTypeErrorID               = 86
	duplicateAttributeErrorID                 = 87
	noFlagsDefinedErrorID                     = 88
	tooManyFlagsDefinedErrorID                = 89
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

type unknownTypeStringCastError struct {
	errorLink
	t Type
}

func newUnknownTypeStringCastError(t Type) *unknownTypeStringCastError {
	return &unknownTypeStringCastError{
		errorLink: errorLink{id: unknownTypeStringCastErrorID},
		t:         t}
}

func (e *unknownTypeStringCastError) Error() string {
	return e.errorf("Unknown type id %q", e.t)
}

type invalidTypeStringCastError struct {
	errorLink
	t Type
}

func newInvalidTypeStringCastError(t Type) *invalidTypeStringCastError {
	return &invalidTypeStringCastError{
		errorLink: errorLink{id: invalidTypeStringCastErrorID},
		t:         t}
}

func (e *invalidTypeStringCastError) Error() string {
	return e.errorf("Can't convert string to value of %q type", e.t)
}

type notImplementedStringCastError struct {
	errorLink
	t Type
}

func newNotImplementedStringCastError(t Type) *notImplementedStringCastError {
	return &notImplementedStringCastError{
		errorLink: errorLink{id: notImplementedStringCastErrorID},
		t:         t}
}

func (e *notImplementedStringCastError) Error() string {
	return e.errorf("Conversion from string to value of %q type hasn't been implemented", e.t)
}

type invalidBooleanStringCastError struct {
	errorLink
	s   string
	err error
}

func newInvalidBooleanStringCastError(s string, err error) *invalidBooleanStringCastError {
	return &invalidBooleanStringCastError{
		errorLink: errorLink{id: invalidBooleanStringCastErrorID},
		s:         s,
		err:       err}
}

func (e *invalidBooleanStringCastError) Error() string {
	return e.errorf("Can't treat %q as boolean (%s)", e.s, e.err)
}

type invalidIntegerStringCastError struct {
	errorLink
	s   string
	err error
}

func newInvalidIntegerStringCastError(s string, err error) *invalidIntegerStringCastError {
	return &invalidIntegerStringCastError{
		errorLink: errorLink{id: invalidIntegerStringCastErrorID},
		s:         s,
		err:       err}
}

func (e *invalidIntegerStringCastError) Error() string {
	return e.errorf("Can't treat %q as integer (%s)", e.s, e.err)
}

type invalidFloatStringCastError struct {
	errorLink
	s   string
	err error
}

func newInvalidFloatStringCastError(s string, err error) *invalidFloatStringCastError {
	return &invalidFloatStringCastError{
		errorLink: errorLink{id: invalidFloatStringCastErrorID},
		s:         s,
		err:       err}
}

func (e *invalidFloatStringCastError) Error() string {
	return e.errorf("Can't treat %q as float (%s)", e.s, e.err)
}

type invalidAddressStringCastError struct {
	errorLink
	s string
}

func newInvalidAddressStringCastError(s string) *invalidAddressStringCastError {
	return &invalidAddressStringCastError{
		errorLink: errorLink{id: invalidAddressStringCastErrorID},
		s:         s}
}

func (e *invalidAddressStringCastError) Error() string {
	return e.errorf("Can't treat %q as IP address", e.s)
}

type invalidNetworkStringCastError struct {
	errorLink
	s   string
	err error
}

func newInvalidNetworkStringCastError(s string, err error) *invalidNetworkStringCastError {
	return &invalidNetworkStringCastError{
		errorLink: errorLink{id: invalidNetworkStringCastErrorID},
		s:         s,
		err:       err}
}

func (e *invalidNetworkStringCastError) Error() string {
	return e.errorf("Can't treat %q as network address (%s)", e.s, e.err)
}

type invalidAddressNetworkStringCastError struct {
	errorLink
	s   string
	err error
}

func newInvalidAddressNetworkStringCastError(s string, err error) *invalidAddressNetworkStringCastError {
	return &invalidAddressNetworkStringCastError{
		errorLink: errorLink{id: invalidAddressNetworkStringCastErrorID},
		s:         s,
		err:       err}
}

func (e *invalidAddressNetworkStringCastError) Error() string {
	return e.errorf("Can't treat %q as address or network (%s)", e.s, e.err)
}

type invalidDomainNameStringCastError struct {
	errorLink
	s   string
	err error
}

func newInvalidDomainNameStringCastError(s string, err error) *invalidDomainNameStringCastError {
	return &invalidDomainNameStringCastError{
		errorLink: errorLink{id: invalidDomainNameStringCastErrorID},
		s:         s,
		err:       err}
}

func (e *invalidDomainNameStringCastError) Error() string {
	return e.errorf("Can't treat %q as domain name (%s)", e.s, e.err)
}

type attributeValueTypeError struct {
	errorLink
	expected Type
	actual   Type
}

func newAttributeValueTypeError(expected, actual Type) *attributeValueTypeError {
	return &attributeValueTypeError{
		errorLink: errorLink{id: attributeValueTypeErrorID},
		expected:  expected,
		actual:    actual}
}

func (e *attributeValueTypeError) Error() string {
	return e.errorf("Expected %q value but got %q", e.expected, e.actual)
}

type attributeValueFlagsTypeError struct {
	errorLink
	t Type
	n int
}

func newAttributeValueFlagsTypeError(t Type, n int) *attributeValueFlagsTypeError {
	return &attributeValueFlagsTypeError{
		errorLink: errorLink{id: attributeValueFlagsTypeErrorID},
		t:         t,
		n:         n}
}

func (e *attributeValueFlagsTypeError) Error() string {
	return e.errorf("Expected %d bits flags value but got %q", e.n, e.t)
}

type attributeValueFlagsBitsError struct {
	errorLink
	t        Type
	expected int
	actual   int
}

func newAttributeValueFlagsBitsError(t Type, expected, actual int) *attributeValueFlagsBitsError {
	return &attributeValueFlagsBitsError{
		errorLink: errorLink{id: attributeValueFlagsBitsErrorID},
		t:         t,
		expected:  expected,
		actual:    actual}
}

func (e *attributeValueFlagsBitsError) Error() string {
	return e.errorf("Expected %d bits flags value but got %q value with %d positions", e.expected, e.t, e.actual)
}

type duplicateAttributeValueError struct {
	errorLink
	ID   string
	t    Type
	curr AttributeValue
	prev AttributeValue
}

func newDuplicateAttributeValueError(ID string, t Type, curr, prev AttributeValue) *duplicateAttributeValueError {
	return &duplicateAttributeValueError{
		errorLink: errorLink{id: duplicateAttributeValueErrorID},
		ID:        ID,
		t:         t,
		curr:      curr,
		prev:      prev}
}

func (e *duplicateAttributeValueError) Error() string {
	return e.errorf("Duplicate attribute %q of type %q in request %s - %s", e.ID, e.t, e.curr.describe(), e.prev.describe())
}

type unknownTypeSerializationError struct {
	errorLink
	t Type
}

func newUnknownTypeSerializationError(t Type) *unknownTypeSerializationError {
	return &unknownTypeSerializationError{
		errorLink: errorLink{id: unknownTypeSerializationErrorID},
		t:         t}
}

func (e *unknownTypeSerializationError) Error() string {
	return e.errorf("Unknown type id %q", e.t)
}

type invalidTypeSerializationError struct {
	errorLink
	t Type
}

func newInvalidTypeSerializationError(t Type) *invalidTypeSerializationError {
	return &invalidTypeSerializationError{
		errorLink: errorLink{id: invalidTypeSerializationErrorID},
		t:         t}
}

func (e *invalidTypeSerializationError) Error() string {
	return e.errorf("Can't serialize value of %q type", e.t)
}

type assignmentTypeMismatch struct {
	errorLink
	a Attribute
	t Type
}

func newAssignmentTypeMismatch(a Attribute, t Type) *assignmentTypeMismatch {
	return &assignmentTypeMismatch{
		errorLink: errorLink{id: assignmentTypeMismatchID},
		a:         a,
		t:         t}
}

func (e *assignmentTypeMismatch) Error() string {
	return e.errorf("Can't assign %q value to attribute %q of type %q", e.t, e.a.id, e.a.t)
}

type mapperArgumentTypeError struct {
	errorLink
	actual Type
}

func newMapperArgumentTypeError(actual Type) *mapperArgumentTypeError {
	return &mapperArgumentTypeError{
		errorLink: errorLink{id: mapperArgumentTypeErrorID},
		actual:    actual}
}

func (e *mapperArgumentTypeError) Error() string {
	return e.errorf("Expected %s, %s or %s as argument but got %s", TypeString, TypeSetOfStrings, TypeListOfStrings, e.actual)
}

// UntaggedPolicyModificationError indicates attempt to modify incrementally a policy which has no tag.
type UntaggedPolicyModificationError struct {
	errorLink
}

func newUntaggedPolicyModificationError() *UntaggedPolicyModificationError {
	return &UntaggedPolicyModificationError{
		errorLink: errorLink{id: UntaggedPolicyModificationErrorID}}
}

// Error implements error interface.
func (e *UntaggedPolicyModificationError) Error() string {
	return e.errorf("Can't modify policies with no tag")
}

// MissingPolicyTagError indicates that update has no tag to match policy before modification.
type MissingPolicyTagError struct {
	errorLink
}

func newMissingPolicyTagError() *MissingPolicyTagError {
	return &MissingPolicyTagError{
		errorLink: errorLink{id: MissingPolicyTagErrorID}}
}

// Error implements error interface.
func (e *MissingPolicyTagError) Error() string {
	return e.errorf("Update has no previous policy tag")
}

// PolicyTagsNotMatchError indicates that update tag doesn't match policy before modification.
type PolicyTagsNotMatchError struct {
	errorLink
	cntTag *uuid.UUID
	updTag *uuid.UUID
}

func newPolicyTagsNotMatchError(cntTag, updTag *uuid.UUID) *PolicyTagsNotMatchError {
	return &PolicyTagsNotMatchError{
		errorLink: errorLink{id: PolicyTagsNotMatchErrorID},
		cntTag:    cntTag,
		updTag:    updTag}
}

// Error implements error interface.
func (e *PolicyTagsNotMatchError) Error() string {
	return e.errorf("Update tag %s doesn't match policies tag %s", e.updTag.String(), e.cntTag.String())
}

type emptyPathModificationError struct {
	errorLink
}

func newEmptyPathModificationError() *emptyPathModificationError {
	return &emptyPathModificationError{
		errorLink: errorLink{id: emptyPathModificationErrorID}}
}

func (e *emptyPathModificationError) Error() string {
	return e.errorf("Can't modify items by empty path")
}

type invalidRootPolicyItemTypeError struct {
	errorLink
	item interface{}
}

func newInvalidRootPolicyItemTypeError(item interface{}) *invalidRootPolicyItemTypeError {
	return &invalidRootPolicyItemTypeError{
		errorLink: errorLink{id: invalidRootPolicyItemTypeErrorID},
		item:      item}
}

func (e *invalidRootPolicyItemTypeError) Error() string {
	return e.errorf("Expected policy or policy set as new root policy but got %T", e.item)
}

type hiddenRootPolicyAppendError struct {
	errorLink
}

func newHiddenRootPolicyAppendError() *hiddenRootPolicyAppendError {
	return &hiddenRootPolicyAppendError{
		errorLink: errorLink{id: hiddenRootPolicyAppendErrorID}}
}

func (e *hiddenRootPolicyAppendError) Error() string {
	return e.errorf("Can't append hidden policy to as root policy")
}

type invalidRootPolicyError struct {
	errorLink
	actual   string
	expected string
}

func newInvalidRootPolicyError(actual, expected string) *invalidRootPolicyError {
	return &invalidRootPolicyError{
		errorLink: errorLink{id: invalidRootPolicyErrorID},
		actual:    actual,
		expected:  expected}
}

func (e *invalidRootPolicyError) Error() string {
	return e.errorf("Root policy is %q but got %q as first path element", e.expected, e.actual)
}

type hiddenPolicySetModificationError struct {
	errorLink
}

func newHiddenPolicySetModificationError() *hiddenPolicySetModificationError {
	return &hiddenPolicySetModificationError{
		errorLink: errorLink{id: hiddenPolicySetModificationErrorID}}
}

func (e *hiddenPolicySetModificationError) Error() string {
	return e.errorf("Can't modify hidden policy set")
}

type invalidPolicySetItemTypeError struct {
	errorLink
	item interface{}
}

func newInvalidPolicySetItemTypeError(item interface{}) *invalidPolicySetItemTypeError {
	return &invalidPolicySetItemTypeError{
		errorLink: errorLink{id: invalidPolicySetItemTypeErrorID},
		item:      item}
}

func (e *invalidPolicySetItemTypeError) Error() string {
	return e.errorf("Expected policy or policy set to append but got %T", e.item)
}

type tooShortPathPolicySetModificationError struct {
	errorLink
}

func newTooShortPathPolicySetModificationError() *tooShortPathPolicySetModificationError {
	return &tooShortPathPolicySetModificationError{
		errorLink: errorLink{id: tooShortPathPolicySetModificationErrorID}}
}

func (e *tooShortPathPolicySetModificationError) Error() string {
	return e.errorf("Path to item to delete is too short")
}

type missingPolicySetChildError struct {
	errorLink
	ID string
}

func newMissingPolicySetChildError(ID string) *missingPolicySetChildError {
	return &missingPolicySetChildError{
		errorLink: errorLink{id: missingPolicySetChildErrorID},
		ID:        ID}
}

func (e *missingPolicySetChildError) Error() string {
	return e.errorf("Policy set has no child policy or policy set with id %q", e.ID)
}

type hiddenPolicyAppendError struct {
	errorLink
}

func newHiddenPolicyAppendError() *hiddenPolicyAppendError {
	return &hiddenPolicyAppendError{
		errorLink: errorLink{id: hiddenPolicyAppendErrorID}}
}

func (e *hiddenPolicyAppendError) Error() string {
	return e.errorf("Can't append hidden policy or policy set")
}

type policyTransactionTagsNotMatchError struct {
	errorLink
	tTag uuid.UUID
	uTag uuid.UUID
}

func newPolicyTransactionTagsNotMatchError(tTag, uTag uuid.UUID) *policyTransactionTagsNotMatchError {
	return &policyTransactionTagsNotMatchError{
		errorLink: errorLink{id: policyTransactionTagsNotMatchErrorID},
		tTag:      tTag,
		uTag:      uTag}
}

func (e *policyTransactionTagsNotMatchError) Error() string {
	return e.errorf("Update tag %s doesn't match policies transaction tag %s", e.uTag.String(), e.tTag.String())
}

type failedPolicyTransactionError struct {
	errorLink
	t   uuid.UUID
	err error
}

func newFailedPolicyTransactionError(t uuid.UUID, err error) *failedPolicyTransactionError {
	return &failedPolicyTransactionError{
		errorLink: errorLink{id: failedPolicyTransactionErrorID},
		t:         t,
		err:       err}
}

func (e *failedPolicyTransactionError) Error() string {
	return e.errorf("Can't operate with failed transaction on policies %s. Previous failure %s", e.t, e.err)
}

type unknownPolicyUpdateOperationError struct {
	errorLink
	op int
}

func newUnknownPolicyUpdateOperationError(op int) *unknownPolicyUpdateOperationError {
	return &unknownPolicyUpdateOperationError{
		errorLink: errorLink{id: unknownPolicyUpdateOperationErrorID},
		op:        op}
}

func (e *unknownPolicyUpdateOperationError) Error() string {
	return e.errorf("Unknown operation %d", e.op)
}

type hiddenPolicyModificationError struct {
	errorLink
}

func newHiddenPolicyModificationError() *hiddenPolicyModificationError {
	return &hiddenPolicyModificationError{
		errorLink: errorLink{id: hiddenPolicyModificationErrorID}}
}

func (e *hiddenPolicyModificationError) Error() string {
	return e.errorf("Can't modify hidden policy")
}

type tooLongPathPolicyModificationError struct {
	errorLink
	path []string
}

func newTooLongPathPolicyModificationError(path []string) *tooLongPathPolicyModificationError {
	return &tooLongPathPolicyModificationError{
		errorLink: errorLink{id: tooLongPathPolicyModificationErrorID},
		path:      path}
}

func (e *tooLongPathPolicyModificationError) Error() string {
	return e.errorf("Trailing path \"%s\"", strings.Join(e.path, "/"))
}

type tooShortPathPolicyModificationError struct {
	errorLink
}

func newTooShortPathPolicyModificationError() *tooShortPathPolicyModificationError {
	return &tooShortPathPolicyModificationError{
		errorLink: errorLink{id: tooShortPathPolicyModificationErrorID}}
}

func (e *tooShortPathPolicyModificationError) Error() string {
	return e.errorf("Path to item to delete is too short")
}

type invalidPolicyItemTypeError struct {
	errorLink
	item interface{}
}

func newInvalidPolicyItemTypeError(item interface{}) *invalidPolicyItemTypeError {
	return &invalidPolicyItemTypeError{
		errorLink: errorLink{id: invalidPolicyItemTypeErrorID},
		item:      item}
}

func (e *invalidPolicyItemTypeError) Error() string {
	return e.errorf("Expected rule to append but got %T", e.item)
}

type hiddenRuleAppendError struct {
	errorLink
}

func newHiddenRuleAppendError() *hiddenRuleAppendError {
	return &hiddenRuleAppendError{
		errorLink: errorLink{id: hiddenRuleAppendErrorID}}
}

func (e *hiddenRuleAppendError) Error() string {
	return e.errorf("Can't append hidden rule")
}

type missingPolicyChildError struct {
	errorLink
	ID string
}

func newMissingPolicyChildError(ID string) *missingPolicyChildError {
	return &missingPolicyChildError{
		errorLink: errorLink{id: missingPolicyChildErrorID},
		ID:        ID}
}

func (e *missingPolicyChildError) Error() string {
	return e.errorf("Policy has no rule with id %q", e.ID)
}

type missingContentError struct {
	errorLink
	ID string
}

func newMissingContentError(ID string) *missingContentError {
	return &missingContentError{
		errorLink: errorLink{id: missingContentErrorID},
		ID:        ID}
}

func (e *missingContentError) Error() string {
	return e.errorf("Missing content %s", e.ID)
}

type invalidContentStorageItem struct {
	errorLink
	ID string
	v  interface{}
}

func newInvalidContentStorageItem(ID string, v interface{}) *invalidContentStorageItem {
	return &invalidContentStorageItem{
		errorLink: errorLink{id: invalidContentStorageItemID},
		ID:        ID,
		v:         v}
}

func (e *invalidContentStorageItem) Error() string {
	return e.errorf("Invalid value at %s (expected *LocalContent but got %T)", e.ID, e.v)
}

type missingContentItemError struct {
	errorLink
	ID string
}

func newMissingContentItemError(ID string) *missingContentItemError {
	return &missingContentItemError{
		errorLink: errorLink{id: missingContentItemErrorID},
		ID:        ID}
}

func (e *missingContentItemError) Error() string {
	return e.errorf("Missing content item %q", e.ID)
}

type invalidContentItemError struct {
	errorLink
	v interface{}
}

func newInvalidContentItemError(v interface{}) *invalidContentItemError {
	return &invalidContentItemError{
		errorLink: errorLink{id: invalidContentItemErrorID},
		v:         v}
}

func (e *invalidContentItemError) Error() string {
	return e.errorf("Invalid value (expected *ContentItem but got %T)", e.v)
}

type invalidContentItemTypeError struct {
	errorLink
	expected Type
	actual   Type
}

func newInvalidContentItemTypeError(expected, actual Type) *invalidContentItemTypeError {
	return &invalidContentItemTypeError{
		errorLink: errorLink{id: invalidContentItemTypeErrorID},
		expected:  expected,
		actual:    actual}
}

func (e *invalidContentItemTypeError) Error() string {
	return e.errorf("Invalid conent item type. Expected %q but got %q", e.expected, e.actual)
}

type invalidSelectorPathError struct {
	errorLink
	expected Signature
	actual   []Expression
}

func newInvalidSelectorPathError(expected Signature, actual []Expression) *invalidSelectorPathError {
	return &invalidSelectorPathError{
		errorLink: errorLink{id: invalidSelectorPathErrorID},
		expected:  expected,
		actual:    actual}
}

func (e *invalidSelectorPathError) Error() string {
	actual := "nothing"
	if len(e.actual) > 0 {
		strs := make([]string, len(e.actual))
		for i, e := range e.actual {
			strs[i] = e.GetResultType().String()
		}
		actual = strings.Join(strs, "/")
	}

	return e.errorf("Invalid selector path. Expected %s path but got %s", e.expected, actual)
}

type networkMapKeyValueTypeError struct {
	errorLink
	t Type
}

func newNetworkMapKeyValueTypeError(t Type) *networkMapKeyValueTypeError {
	return &networkMapKeyValueTypeError{
		errorLink: errorLink{id: networkMapKeyValueTypeErrorID},
		t:         t}
}

func (e *networkMapKeyValueTypeError) Error() string {
	return e.errorf("Expected %q or %q as network map key but got %q", TypeAddress, TypeNetwork, e.t)
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

type invalidContentModificationError struct {
	errorLink
}

func newInvalidContentModificationError() *invalidContentModificationError {
	return &invalidContentModificationError{
		errorLink: errorLink{id: invalidContentModificationErrorID}}
}

func (e *invalidContentModificationError) Error() string {
	return e.errorf("Can't modify non-mapping content item")
}

type missingPathContentModificationError struct {
	errorLink
}

func newMissingPathContentModificationError() *missingPathContentModificationError {
	return &missingPathContentModificationError{
		errorLink: errorLink{id: missingPathContentModificationErrorID}}
}

func (e *missingPathContentModificationError) Error() string {
	return e.errorf("Missing path for content item change")
}

type tooLongPathContentModificationError struct {
	errorLink
	expected Signature
	actual   []AttributeValue
}

func newTooLongPathContentModificationError(expected Signature, actual []AttributeValue) *tooLongPathContentModificationError {
	return &tooLongPathContentModificationError{
		errorLink: errorLink{id: tooLongPathContentModificationErrorID},
		expected:  expected,
		actual:    actual}
}

func (e *tooLongPathContentModificationError) Error() string {
	actStrs := make([]string, len(e.actual))
	for i, e := range e.actual {
		actStrs[i] = strconv.Quote(e.GetResultType().String())
	}
	actual := strings.Join(actStrs, "/")

	return e.errorf("Too long modification path. Expected %s path but got %s", e.expected, actual)
}

type invalidContentValueModificationError struct {
	errorLink
}

func newInvalidContentValueModificationError() *invalidContentValueModificationError {
	return &invalidContentValueModificationError{
		errorLink: errorLink{id: invalidContentValueModificationErrorID}}
}

func (e *invalidContentValueModificationError) Error() string {
	return e.errorf("Can't modify final content value")
}

// UntaggedContentModificationError indicates attempt to modify incrementally a content which has no tag.
type UntaggedContentModificationError struct {
	errorLink
	ID string
}

func newUntaggedContentModificationError(ID string) *UntaggedContentModificationError {
	return &UntaggedContentModificationError{
		errorLink: errorLink{id: UntaggedContentModificationErrorID},
		ID:        ID}
}

// Error implements error interface.
func (e *UntaggedContentModificationError) Error() string {
	return e.errorf("Can't modify content %q with no tag", e.ID)
}

// MissingContentTagError indicates that update has no tag to match content before modification.
type MissingContentTagError struct {
	errorLink
}

func newMissingContentTagError() *MissingContentTagError {
	return &MissingContentTagError{
		errorLink: errorLink{id: MissingContentTagErrorID}}
}

// Error implements error interface.
func (e *MissingContentTagError) Error() string {
	return e.errorf("Update has no previous content tag")
}

// ContentTagsNotMatchError indicates that update tag doesn't match content before modification.
type ContentTagsNotMatchError struct {
	errorLink
	ID     string
	cntTag *uuid.UUID
	updTag *uuid.UUID
}

func newContentTagsNotMatchError(ID string, cntTag, updTag *uuid.UUID) *ContentTagsNotMatchError {
	return &ContentTagsNotMatchError{
		errorLink: errorLink{id: ContentTagsNotMatchErrorID},
		ID:        ID,
		cntTag:    cntTag,
		updTag:    updTag}
}

// Error implements error interface.
func (e *ContentTagsNotMatchError) Error() string {
	return e.errorf("Update tag %s doesn't match content %q tag %s", e.cntTag.String(), e.ID, e.updTag.String())
}

type unknownContentUpdateOperationError struct {
	errorLink
	op int
}

func newUnknownContentUpdateOperationError(op int) *unknownContentUpdateOperationError {
	return &unknownContentUpdateOperationError{
		errorLink: errorLink{id: unknownContentUpdateOperationErrorID},
		op:        op}
}

func (e *unknownContentUpdateOperationError) Error() string {
	return e.errorf("Unknown operation %d", e.op)
}

type failedContentTransactionError struct {
	errorLink
	id  string
	t   uuid.UUID
	err error
}

func newFailedContentTransactionError(id string, t uuid.UUID, err error) *failedContentTransactionError {
	return &failedContentTransactionError{
		errorLink: errorLink{id: failedContentTransactionErrorID},
		id:        id,
		t:         t,
		err:       err}
}

func (e *failedContentTransactionError) Error() string {
	return e.errorf("Can't operate with failed transaction on content %q tagged %s. Previous failure %s", e.id, e.t, e.err)
}

type contentTransactionIDNotMatchError struct {
	errorLink
	tID string
	uID string
}

func newContentTransactionIDNotMatchError(tID, uID string) *contentTransactionIDNotMatchError {
	return &contentTransactionIDNotMatchError{
		errorLink: errorLink{id: contentTransactionIDNotMatchErrorID},
		tID:       tID,
		uID:       uID}
}

func (e *contentTransactionIDNotMatchError) Error() string {
	return e.errorf("Update content ID %q doesn't match transaction content ID %q", e.uID, e.tID)
}

type contentTransactionTagsNotMatchError struct {
	errorLink
	id   string
	tTag uuid.UUID
	uTag uuid.UUID
}

func newContentTransactionTagsNotMatchError(id string, tTag, uTag uuid.UUID) *contentTransactionTagsNotMatchError {
	return &contentTransactionTagsNotMatchError{
		errorLink: errorLink{id: contentTransactionTagsNotMatchErrorID},
		id:        id,
		tTag:      tTag,
		uTag:      uTag}
}

func (e *contentTransactionTagsNotMatchError) Error() string {
	return e.errorf("Update tag %s doesn't match content %q transaction tag %s", e.uTag.String(), e.id, e.tTag.String())
}

type tooShortRawPathContentModificationError struct {
	errorLink
}

func newTooShortRawPathContentModificationError() *tooShortRawPathContentModificationError {
	return &tooShortRawPathContentModificationError{
		errorLink: errorLink{id: tooShortRawPathContentModificationErrorID}}
}

func (e *tooShortRawPathContentModificationError) Error() string {
	return e.errorf("Expected at least content item ID in path but got nothing")
}

type tooLongRawPathContentModificationError struct {
	errorLink
	expected Signature
	actual   []string
}

func newTooLongRawPathContentModificationError(expected Signature, actual []string) *tooLongRawPathContentModificationError {
	return &tooLongRawPathContentModificationError{
		errorLink: errorLink{id: tooLongRawPathContentModificationErrorID},
		expected:  expected,
		actual:    actual}
}

func (e *tooLongRawPathContentModificationError) Error() string {
	actStrs := make([]string, len(e.actual))
	for i, s := range e.actual {
		actStrs[i] = strconv.Quote(s)
	}
	actual := strings.Join(actStrs, "/")

	return e.errorf("Too long modification path. Expected %s path but got %s", e.expected, actual)
}

type invalidContentUpdateDataError struct {
	errorLink
	v interface{}
}

func newInvalidContentUpdateDataError(v interface{}) *invalidContentUpdateDataError {
	return &invalidContentUpdateDataError{
		errorLink: errorLink{id: invalidContentUpdateDataErrorID},
		v:         v}
}

func (e *invalidContentUpdateDataError) Error() string {
	return e.errorf("Expected content update data but got %T", e.v)
}

type invalidContentUpdateResultTypeError struct {
	errorLink
	actual   Type
	expected Type
}

func newInvalidContentUpdateResultTypeError(actual, expected Type) *invalidContentUpdateResultTypeError {
	return &invalidContentUpdateResultTypeError{
		errorLink: errorLink{id: invalidContentUpdateResultTypeErrorID},
		actual:    actual,
		expected:  expected}
}

func (e *invalidContentUpdateResultTypeError) Error() string {
	return e.errorf("Expected %q as a result type but got %q", e.expected, e.actual)
}

type invalidContentUpdateKeysError struct {
	errorLink
	start    int
	actual   Signature
	expected Signature
}

func newInvalidContentUpdateKeysError(start int, actual, expected Signature) *invalidContentUpdateKeysError {
	return &invalidContentUpdateKeysError{
		errorLink: errorLink{id: invalidContentUpdateKeysErrorID},
		start:     start,
		actual:    actual,
		expected:  expected}
}

func (e *invalidContentUpdateKeysError) Error() string {
	return e.errorf("Expected %s path after position %d but got %s", e.expected, e.start, e.actual)
}

type unknownContentItemResultTypeError struct {
	errorLink
	t Type
}

func newUnknownContentItemResultTypeError(t Type) *unknownContentItemResultTypeError {
	return &unknownContentItemResultTypeError{
		errorLink: errorLink{id: unknownContentItemResultTypeErrorID},
		t:         t}
}

func (e *unknownContentItemResultTypeError) Error() string {
	return e.errorf("Unknown result type for content item: %q", e.t)
}

type invalidContentItemResultTypeError struct {
	errorLink
	t Type
}

func newInvalidContentItemResultTypeError(t Type) *invalidContentItemResultTypeError {
	return &invalidContentItemResultTypeError{
		errorLink: errorLink{id: invalidContentItemResultTypeErrorID},
		t:         t}
}

func (e *invalidContentItemResultTypeError) Error() string {
	return e.errorf("Invalid result type for content item: %q", e.t)
}

type invalidContentKeyTypeError struct {
	errorLink
	t        Type
	expected TypeSet
}

func newInvalidContentKeyTypeError(t Type, expected TypeSet) *invalidContentKeyTypeError {
	return &invalidContentKeyTypeError{
		errorLink: errorLink{id: invalidContentKeyTypeErrorID},
		t:         t,
		expected:  expected}
}

func (e *invalidContentKeyTypeError) Error() string {
	names := make([]string, len(e.expected))
	i := 0
	for t := range e.expected {
		names[i] = strconv.Quote(t.String())
		i++
	}
	s := strings.Join(names, ", ")

	return e.errorf("Invalid key type for content item: %q (expected %s)", e.t, s)
}

type invalidContentStringMapError struct {
	errorLink
	value interface{}
}

func newInvalidContentStringMapError(value interface{}) *invalidContentStringMapError {
	return &invalidContentStringMapError{
		errorLink: errorLink{id: invalidContentStringMapErrorID},
		value:     value}
}

func (e *invalidContentStringMapError) Error() string {
	return e.errorf("Expected string map but got %T", e.value)
}

type invalidContentNetworkMapError struct {
	errorLink
	value interface{}
}

func newInvalidContentNetworkMapError(value interface{}) *invalidContentNetworkMapError {
	return &invalidContentNetworkMapError{
		errorLink: errorLink{id: invalidContentNetworkMapErrorID},
		value:     value}
}

func (e *invalidContentNetworkMapError) Error() string {
	return e.errorf("Expected network map but got %T", e.value)
}

type invalidContentDomainMapError struct {
	errorLink
	value interface{}
}

func newInvalidContentDomainMapError(value interface{}) *invalidContentDomainMapError {
	return &invalidContentDomainMapError{
		errorLink: errorLink{id: invalidContentDomainMapErrorID},
		value:     value}
}

func (e *invalidContentDomainMapError) Error() string {
	return e.errorf("Expected domain map but got %T", e.value)
}

type invalidContentValueError struct {
	errorLink
	value interface{}
}

func newInvalidContentValueError(value interface{}) *invalidContentValueError {
	return &invalidContentValueError{
		errorLink: errorLink{id: invalidContentValueErrorID},
		value:     value}
}

func (e *invalidContentValueError) Error() string {
	return e.errorf("Expected value but got %T", e.value)
}

type invalidContentValueTypeError struct {
	errorLink
	value    interface{}
	expected Type
}

func newInvalidContentValueTypeError(value interface{}, expected Type) *invalidContentValueTypeError {
	return &invalidContentValueTypeError{
		errorLink: errorLink{id: invalidContentValueTypeErrorID},
		value:     value,
		expected:  expected}
}

func (e *invalidContentValueTypeError) Error() string {
	return e.errorf("Expected value of type %q but got %T", e.expected, e.value)
}

type integerDivideByZeroError struct {
	errorLink
}

func newIntegerDivideByZeroError() *integerDivideByZeroError {
	return &integerDivideByZeroError{
		errorLink: errorLink{id: integerDivideByZeroErrorID}}
}

func (e *integerDivideByZeroError) Error() string {
	return e.errorf("Integer divisor has a value of 0")
}

type floatDivideByZeroError struct {
	errorLink
}

func newFloatDivideByZeroError() *floatDivideByZeroError {
	return &floatDivideByZeroError{
		errorLink: errorLink{id: floatDivideByZeroErrorID}}
}

func (e *floatDivideByZeroError) Error() string {
	return e.errorf("Float divisor has a value of 0")
}

type floatNanError struct {
	errorLink
}

func newFloatNanError() *floatNanError {
	return &floatNanError{
		errorLink: errorLink{id: floatNanErrorID}}
}

func (e *floatNanError) Error() string {
	return e.errorf("Float result has a value of NaN")
}

type floatInfError struct {
	errorLink
}

func newFloatInfError() *floatInfError {
	return &floatInfError{
		errorLink: errorLink{id: floatInfErrorID}}
}

func (e *floatInfError) Error() string {
	return e.errorf("Float result has a value of Inf")
}

type nilTypeError struct {
	errorLink
}

func newNilTypeError() *nilTypeError {
	return &nilTypeError{
		errorLink: errorLink{id: nilTypeErrorID}}
}

func (e *nilTypeError) Error() string {
	return e.errorf("Can't put nil type into custom types symbol table")
}

type builtinCustomTypeError struct {
	errorLink
	t Type
}

func newBuiltinCustomTypeError(t Type) *builtinCustomTypeError {
	return &builtinCustomTypeError{
		errorLink: errorLink{id: builtinCustomTypeErrorID},
		t:         t}
}

func (e *builtinCustomTypeError) Error() string {
	return e.errorf("Can't put built-in type %q into custom types symbol table", e.t)
}

type duplicateCustomTypeError struct {
	errorLink
	n Type
	p Type
}

func newDuplicateCustomTypeError(n, p Type) *duplicateCustomTypeError {
	return &duplicateCustomTypeError{
		errorLink: errorLink{id: duplicateCustomTypeErrorID},
		n:         n,
		p:         p}
}

func (e *duplicateCustomTypeError) Error() string {
	return e.errorf("Can't put type %q into symbol table as it already contains %q", e.n, e.p)
}

type duplicatesBuiltinTypeError struct {
	errorLink
	name string
}

func newDuplicatesBuiltinTypeError(name string) *duplicatesBuiltinTypeError {
	return &duplicatesBuiltinTypeError{
		errorLink: errorLink{id: duplicatesBuiltinTypeErrorID},
		name:      name}
}

func (e *duplicatesBuiltinTypeError) Error() string {
	return e.errorf("Can't create flags type %q. The name is taken by a built-in types", e.name)
}

type duplicateFlagName struct {
	errorLink
	name string
	flag string
	i    int
	j    int
}

func newDuplicateFlagName(name, flag string, i, j int) *duplicateFlagName {
	return &duplicateFlagName{
		errorLink: errorLink{id: duplicateFlagNameID},
		name:      name,
		flag:      flag,
		i:         i,
		j:         j}
}

func (e *duplicateFlagName) Error() string {
	return e.errorf("Can't create flags type %q. Flag %q at %d position duplicates flag at %d", e.name, e.flag, e.i, e.j)
}

type noTypedAttributeError struct {
	errorLink
	a Attribute
}

func newNoTypedAttributeError(a Attribute) *noTypedAttributeError {
	return &noTypedAttributeError{
		errorLink: errorLink{id: noTypedAttributeErrorID},
		a:         a}
}

func (e *noTypedAttributeError) Error() string {
	return e.errorf("Attribute %q has no type", e.a.id)
}

type undefinedAttributeTypeError struct {
	errorLink
	a Attribute
}

func newUndefinedAttributeTypeError(a Attribute) *undefinedAttributeTypeError {
	return &undefinedAttributeTypeError{
		errorLink: errorLink{id: undefinedAttributeTypeErrorID},
		a:         a}
}

func (e *undefinedAttributeTypeError) Error() string {
	return e.errorf("Attribute %q has type %q", e.a.id, TypeUndefined)
}

type unknownAttributeTypeError struct {
	errorLink
	a Attribute
}

func newUnknownAttributeTypeError(a Attribute) *unknownAttributeTypeError {
	return &unknownAttributeTypeError{
		errorLink: errorLink{id: unknownAttributeTypeErrorID},
		a:         a}
}

func (e *unknownAttributeTypeError) Error() string {
	return e.errorf("Attribute %q has unknown type %q", e.a.id, e.a.t)
}

type duplicateAttributeError struct {
	errorLink
	a Attribute
}

func newDuplicateAttributeError(a Attribute) *duplicateAttributeError {
	return &duplicateAttributeError{
		errorLink: errorLink{id: duplicateAttributeErrorID},
		a:         a}
}

func (e *duplicateAttributeError) Error() string {
	return e.errorf("Can't put attribute %q into symbol table as it already contains one with the same id", e.a.id)
}

type noFlagsDefinedError struct {
	errorLink
	name string
	n    int
}

func newNoFlagsDefinedError(name string, n int) *noFlagsDefinedError {
	return &noFlagsDefinedError{
		errorLink: errorLink{id: noFlagsDefinedErrorID},
		name:      name,
		n:         n}
}

func (e *noFlagsDefinedError) Error() string {
	return e.errorf("Required at least one flag to define flags type %q got %d", e.name, e.n)
}

type tooManyFlagsDefinedError struct {
	errorLink
	name string
	n    int
}

func newTooManyFlagsDefinedError(name string, n int) *tooManyFlagsDefinedError {
	return &tooManyFlagsDefinedError{
		errorLink: errorLink{id: tooManyFlagsDefinedErrorID},
		name:      name,
		n:         n}
}

func (e *tooManyFlagsDefinedError) Error() string {
	return e.errorf("Required no more than 64 flags to define flags type %q got %d", e.name, e.n)
}
