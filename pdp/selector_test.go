package pdp

import (
	"fmt"
	"strings"
	"testing"
)

func TestSelector(t *testing.T) {
	c, v := prepareTestYAST(YASTTestSelector, YASTSelectorTestAttrs, YASTSelectorTestContent, t)

	s, err := c.unmarshalSelector(v)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	if s == nil {
		t.Fatalf("Expected selector but got nothing")
	}

	errors := assertContentHasMissesOrErrors(s.Content)
	if len(errors) > 0 {
		for i, err := range errors {
			t.Errorf("%d - %s", i+1, err)
		}
	}

	ctx := NewContext()
	ctx.StoreAttribute("s", DataTypeString, "test")
	ctx.StoreAttribute("d", DataTypeDomain, "example.com")

	a, err := s.calculate(&ctx)
	if err != nil {
		t.Errorf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	} else {
		assertStringValue(a, "first", t)
	}

	ctx = NewContext()
	ctx.StoreAttribute("s", DataTypeString, "example")
	ctx.StoreAttribute("d", DataTypeDomain, "www.example.net")

	a, err = s.calculate(&ctx)
	if err != nil {
		t.Errorf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	} else {
		assertStringValue(a, "fifth", t)
	}
}

func assertContentHasMissesOrErrors(v interface{}) []error {
	return assertContentElementHasMissesOrErrors(v, []string{}, []error{})
}

func assertContentElementHasMissesOrErrors(v interface{}, path []string, errors []error) []error {
	switch v := v.(type) {
	case error:
		return append(errors, fmt.Errorf("Invalid selector content at /%s: %s", strings.Join(path, "/"), v))

	case missingSelectorValue:
		return append(errors, fmt.Errorf("Content misses required data at /%s: %s", strings.Join(path, "/"), v.err))

	case map[string]interface{}:
		return assertContentMapHasMissesOrErrors(v, path, errors)

	case []interface{}:
		return assertContentArrayHasMissesOrErrors(v, path, errors)

	case *SetOfSubdomains:
		return assertContentSetOfSubDomainsHasMissesOrErrors(v, path, errors)

	case AttributeValueType:
		return errors
	}

	return append(errors, fmt.Errorf("Unknown content element at /%s: %T", strings.Join(path, "/"), v))
}

func assertContentMapHasMissesOrErrors(m map[string]interface{}, path []string, errors []error) []error {
	for k, v := range m {
		errors = assertContentElementHasMissesOrErrors(v, append(path, k), errors)
	}

	return errors
}

func assertContentArrayHasMissesOrErrors(a []interface{}, path []string, errors []error) []error {
	for i, v := range a {
		errors = assertContentElementHasMissesOrErrors(v, append(path, fmt.Sprintf("%d", i)), errors)
	}

	return errors
}

func assertContentSetOfSubDomainsHasMissesOrErrors(s *SetOfSubdomains, path []string, errors []error) []error {
	for v := range s.Iterate() {
		errors = assertContentElementHasMissesOrErrors(v.Leaf, append(path, v.Domain), errors)
	}

	return errors
}

func assertStringValue(a AttributeValueType, s string, t *testing.T) {
	if a.DataType != DataTypeString {
		t.Errorf("Expected %s attribute but got %s", DataTypeNames[DataTypeString], DataTypeNames[a.DataType])
		return
	}

	v := a.Value.(string)
	if v != s {
		t.Errorf("Expected %q but got %q", s, v)
	}
}
