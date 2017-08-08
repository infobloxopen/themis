package pdp

/* AUTOMATICALLY GENERATED FROM errors.yaml - DO NOT EDIT */

import (
	"github.com/satori/go.uuid"
	"strings"
)

const (
	externalErrorID = iota
	multiErrorID
	missingAttributeErrorID
	missingValueErrorID
	attributeValueTypeErrorID
	mapperArgumentTypeErrorID
	untaggedPolicyModificationErrorID
	missingPolicyTagErrorID
	emptyPathModificationErrorID
	invalidRootPolicyItemTypeErrorID
	hiddenRootPolicyAppendErrorID
	invalidRootPolicyErrorID
	hiddenPolicySetModificationErrorID
	invalidPolicySetItemTypeErrorID
	tooShortPathPolicySetModificationErrorID
	missingPolicySetChildErrorID
	hiddenPolicyAppendErrorID
	policyTagsNotMatchErrorID
	hiddenPolicyModificationErrorID
	tooLongPathPolicyModificationErrorID
	tooShortPathPolicyModificationErrorID
	invalidPolicyItemTypeErrorID
	hiddenRuleAppendErrorID
	missingPolicyChildErrorID
	missingContentErrorID
	invalidContentStorageItemID
	missingContentItemErrorID
	invalidContentItemErrorID
	invalidContentItemTypeErrorID
	invalidSelectorPathErrorID
	networkMapKeyValueTypeErrorID
	mapContentSubitemErrorID
	invalidContentModificationErrorID
	missingPathContentModificationErrorID
	tooLongPathContentModificationErrorID
	invalidContentValueModificationErrorID
	untaggedContentModificationErrorID
	missingContentTagErrorID
	contentTagsNotMatchErrorID
	unknownContentItemResultTypeErrorID
	invalidContentItemResultTypeErrorID
	invalidContentKeyTypeErrorID
	invalidContentStringMapErrorID
	invalidContentNetworkMapErrorID
	invalidContentDomainMapErrorID
	invalidContentValueErrorID
	invalidContentValueTypeErrorID
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

type untaggedPolicyModificationError struct {
	errorLink
}

func newUntaggedPolicyModificationError() *untaggedPolicyModificationError {
	return &untaggedPolicyModificationError{
		errorLink: errorLink{id: untaggedPolicyModificationErrorID}}
}

func (e *untaggedPolicyModificationError) Error() string {
	return e.errorf("Can't modify policies with no tag")
}

type missingPolicyTagError struct {
	errorLink
}

func newMissingPolicyTagError() *missingPolicyTagError {
	return &missingPolicyTagError{
		errorLink: errorLink{id: missingPolicyTagErrorID}}
}

func (e *missingPolicyTagError) Error() string {
	return e.errorf("Update has no previous policy tag")
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

type policyTagsNotMatchError struct {
	errorLink
	cntTag *uuid.UUID
	updTag *uuid.UUID
}

func newPolicyTagsNotMatchError(cntTag, updTag *uuid.UUID) *policyTagsNotMatchError {
	return &policyTagsNotMatchError{
		errorLink: errorLink{id: policyTagsNotMatchErrorID},
		cntTag:    cntTag,
		updTag:    updTag}
}

func (e *policyTagsNotMatchError) Error() string {
	return e.errorf("Update tag %s doesn't match policies tag %s", e.cntTag.String(), e.updTag.String())
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
	return e.errorf("Invalid value at %s (expected *localContent but got %T)", e.ID, e.v)
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

type networkMapKeyValueTypeError struct {
	errorLink
	t int
}

func newNetworkMapKeyValueTypeError(t int) *networkMapKeyValueTypeError {
	return &networkMapKeyValueTypeError{
		errorLink: errorLink{id: networkMapKeyValueTypeErrorID},
		t:         t}
}

func (e *networkMapKeyValueTypeError) Error() string {
	return e.errorf("Expected %s or %s as network map key but got %s", TypeNames[TypeAddress], TypeNames[TypeNetwork], TypeNames[e.t])
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
	expected []int
	actual   []AttributeValue
}

func newTooLongPathContentModificationError(expected []int, actual []AttributeValue) *tooLongPathContentModificationError {
	return &tooLongPathContentModificationError{
		errorLink: errorLink{id: tooLongPathContentModificationErrorID},
		expected:  expected,
		actual:    actual}
}

func (e *tooLongPathContentModificationError) Error() string {
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

	return e.errorf("Too long modification path. Expected at most %s but got %s", expected, actual)
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

type untaggedContentModificationError struct {
	errorLink
	ID string
}

func newUntaggedContentModificationError(ID string) *untaggedContentModificationError {
	return &untaggedContentModificationError{
		errorLink: errorLink{id: untaggedContentModificationErrorID},
		ID:        ID}
}

func (e *untaggedContentModificationError) Error() string {
	return e.errorf("Can't modify content %q with no tag", e.ID)
}

type missingContentTagError struct {
	errorLink
}

func newMissingContentTagError() *missingContentTagError {
	return &missingContentTagError{
		errorLink: errorLink{id: missingContentTagErrorID}}
}

func (e *missingContentTagError) Error() string {
	return e.errorf("Update has no previous content tag")
}

type contentTagsNotMatchError struct {
	errorLink
	ID     string
	cntTag *uuid.UUID
	updTag *uuid.UUID
}

func newContentTagsNotMatchError(ID string, cntTag, updTag *uuid.UUID) *contentTagsNotMatchError {
	return &contentTagsNotMatchError{
		errorLink: errorLink{id: contentTagsNotMatchErrorID},
		ID:        ID,
		cntTag:    cntTag,
		updTag:    updTag}
}

func (e *contentTagsNotMatchError) Error() string {
	return e.errorf("Update tag %s doesn't match content %q tag %s", e.cntTag.String(), e.ID, e.updTag.String())
}

type unknownContentItemResultTypeError struct {
	errorLink
	t int
}

func newUnknownContentItemResultTypeError(t int) *unknownContentItemResultTypeError {
	return &unknownContentItemResultTypeError{
		errorLink: errorLink{id: unknownContentItemResultTypeErrorID},
		t:         t}
}

func (e *unknownContentItemResultTypeError) Error() string {
	return e.errorf("Unknown result type for content item: %d", e.t)
}

type invalidContentItemResultTypeError struct {
	errorLink
	t int
}

func newInvalidContentItemResultTypeError(t int) *invalidContentItemResultTypeError {
	return &invalidContentItemResultTypeError{
		errorLink: errorLink{id: invalidContentItemResultTypeErrorID},
		t:         t}
}

func (e *invalidContentItemResultTypeError) Error() string {
	return e.errorf("Invalid result type for content item: %s", TypeNames[e.t])
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
		names[i] = TypeNames[t]
		i++
	}
	s := strings.Join(names, ", ")

	return e.errorf("Invalid key type for content item: %s (expected %s)", TypeNames[e.t], s)
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
	expected int
}

func newInvalidContentValueTypeError(value interface{}, expected int) *invalidContentValueTypeError {
	return &invalidContentValueTypeError{
		errorLink: errorLink{id: invalidContentValueTypeErrorID},
		value:     value,
		expected:  expected}
}

func (e *invalidContentValueTypeError) Error() string {
	return e.errorf("Expected value of type %s but got %T", TypeNames[e.expected], e.value)
}
