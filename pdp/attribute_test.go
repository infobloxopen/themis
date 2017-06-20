package pdp

import (
	"strings"
	"testing"
)

func TestAttributeValueDescribe(t *testing.T) {
	v := AttributeValueType{
		DataType: 2415919103,
		Value:    nil}

	s := v.describe()
	if s != "<unknown value type 2415919103>" {
		t.Errorf("Expected description %q for unknown type but got %q", "<unknown value type 2415919103>", s)
	}

	v = AttributeValueType{
		DataType: DataTypeListOfStrings,
		Value:    []string{"one", "two", "three", "four"}}

	s = v.describe()
	if s != "[one, two, three, four]" {
		t.Errorf("Expected description %q for %s type but got %q",
			"[one, two, three, four]", DataTypeNames[DataTypeListOfStrings], s)
	}
}

func TestAttributeValueExtract(t *testing.T) {
	v := AttributeValueType{
		DataType: DataTypeString,
		Value:    "test"}

	s, err := ExtractListOfStringsValue(v, "example")
	if err == nil {
		t.Errorf("Expected error for exstracting %s value from %s but got result %#v",
			DataTypeNames[DataTypeListOfStrings], DataTypeNames[DataTypeString], s)
	}

	v = AttributeValueType{
		DataType: DataTypeListOfStrings,
		Value:    []string{"one", "two", "three", "four"}}

	s, err = ExtractListOfStringsValue(v, "example")
	if err != nil {
		t.Errorf("Expected no error for exstracting %s value from %s but got %s",
			DataTypeNames[DataTypeListOfStrings], DataTypeNames[DataTypeString], err)
	} else {
		if strings.Join(s, ", ") != "one, two, three, four" {
			t.Errorf("Expected array %q for exstracting %s value from %s but got %#v",
				"one, two, three, four",
				DataTypeNames[DataTypeListOfStrings],
				DataTypeNames[DataTypeString], strings.Join(s, ", "))
		}
	}
}
