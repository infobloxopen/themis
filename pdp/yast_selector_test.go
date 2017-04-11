package pdp

import (
	"strings"
	"testing"
)

const (
	YASTTestSelector = `# Test selector
type: String
path:
  - test
  - attr: s
  - val:
      type: String
      content: example
  - "0"
  - attr: d
  - selector:
      type: String
      path:
        - test
      content: t
content: c`

	YASTTestSelectorDisplayPath = "\"test\"/attr(s)/\"example\"/\"0\"/attr(d)/selector(\"t\":\"test\")"

	YASTTestSelectorWithNetworks = `# Test selector with networks
type: String
path:
  - attr: a
  - attr: "n"
content: c`

	YASTTestInvalidSelector = `# Test invalid selector
- 0
- 1
- 2`

	YASTTestSelectorMissingType = `# Test selector missing type
path:
  - test
  - attr: s
  - val:
      type: String
      content: example
  - attr: d
content: c`

	YASTTestSelectorInvalidTypeType = `# Test selector invalid type of type
type: 0
path:
  - test
  - attr: s
  - val:
      type: String
      content: example
  - attr: d
content: c`

	YASTTestSelectorInvalidTypeValue = `# Test selector invalid value of type
type: Invalid
path:
  - test
  - attr: s
  - val:
      type: String
      content: example
  - attr: d
content: c`

	YASTTestSelectorMissingPath = `# Test selector missing path
type: String
content: c`

	YASTTestSelectorInvalidPathType = `# Test selector invalid path type
type: String
path: Invalid
content: c`

	YASTTestSelectorInvalidPathElementType = `# Test selector invalid path element type
type: String
path:
  - true
content: c`

	YASTTestSelectorInvalidPathElementMap = `# Test selector invalid path element map
type: String
path:
  - x: 0
    y: 1
    z: 2
content: c`

	YASTTestSelectorInvalidPathElementSpecType = `# Test selector invalid path element specificator type
type: String
path:
  - 0: 0
content: c`

	YASTTestSelectorInvalidPathElementSpec = `# Test selector invalid path element specificator
type: String
path:
  - invalid: 0
content: c`

	YASTTestSelectorInvalidPathElementValue = `# Test selector invalid path element value
type: String
path:
  - val: 0
content: c`

	YASTTestSelectorInvalidPathElementValueType = `# Test selector invalid path element value type
type: String
path:
  - val:
      type: network
      content: 127.0.0.1/8
content: c`

	YASTTestSelectorInvalidPathElementAttr = `# Test selector invalid path element attribute
type: String
path:
  - attr: 0
content: c`

	YASTTestSelectorUnknownPathElementAttr = `# Test selector unknown path element attribute
type: String
path:
  - attr: x
content: c`

	YASTTestSelectorInvalidPathElementAttrType = `# Test selector invalid path element attribute type
type: String
path:
  - attr: set
content: c`

	YASTTestSelectorInvalidContent = `# Test selector invalid content
type: String
path:
  - test
  - attr: s
  - val:
      type: String
      content: example
  - attr: d
content: 0`

	YASTTestSelectorUnknownContent = `# Test selector unknown content
type: String
path:
  - test
  - attr: s
  - val:
      type: String
      content: example
  - attr: d
content: x`

	YASTTestSelectorInvalidSubselector = `# Test selector invalid subselector
type: String
path:
  - test
  - attr: s
  - val:
      type: String
      content: example
  - attr: d
  - selector:
      path:
        - test
      content: t
content: c`

	YASTTestSelectorInvalidSubselectorType = `# Test selector invalid subselector type
type: String
path:
  - test
  - attr: s
  - val:
      type: String
      content: example
  - attr: d
  - selector:
      type: boolean
      path:
        - test
      content: t
content: c`
)

var (
	YASTSelectorTestAttrs = map[string]AttributeType{
		"s":   AttributeType{ID: "s", DataType: DataTypeString},
		"d":   AttributeType{ID: "d", DataType: DataTypeDomain},
		"a":   AttributeType{ID: "a", DataType: DataTypeAddress},
		"n":   AttributeType{ID: "n", DataType: DataTypeNetwork},
		"set": AttributeType{ID: "set", DataType: DataTypeSetOfNetworks}}

	YASTSelectorTestContent = map[string]interface{}{
		"c": map[string]interface{}{
			"test": map[string]interface{}{
				"test": map[string]interface{}{
					"example": []interface{}{
						map[string]interface{}{
							"example.com": map[string]interface{}{
								"example": "first"},
							"www.test.com": map[string]interface{}{
								"example": "second"},
							"wiki.test.com": map[string]interface{}{
								"example": "third"}},
						"unreacheable"},
					"test": "unreacheable"},
				"example": map[string]interface{}{
					"example": []interface{}{
						map[string]interface{}{
							"test.net": map[string]interface{}{
								"example": "fourth"},
							"www.example.net": map[string]interface{}{
								"example": "fifth"},
							"mail.example.net": map[string]interface{}{
								"example": "sixth"}},
						"unreacheable"},
					"test": "unreacheable"}},
			"example": "unreacheable"},
		"t": map[string]interface{}{
			"test": "example"}}

	YASTSelectorTestContentWithNetworks = map[string]interface{}{
		"c": map[string]interface{}{
			"10.0.0.1/8": map[string]interface{}{
				"10.0.0.1/16": "first",
				"10.1.0.1/16": "second"},
			"192.168.0.1/16": map[string]interface{}{
				"192.168.0.1/24": "third",
				"192.168.1.0/24": "fourth"},
			"127.0.0.1/8": map[string]interface{}{
				"127.0.0.1/24": "fifth",
				"127.0.1.1/24": "sixth"}}}
)

func TestUnmarshalYASTSelector(t *testing.T) {
	c, v := prepareTestYAST(YASTTestSelector, YASTSelectorTestAttrs, YASTSelectorTestContent, t)

	s, err := c.unmarshalSelector(v)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	if s == nil {
		t.Fatalf("Expected selector but got nothing")
	}

	if s.DataType != DataTypeString {
		dataType, ok := DataTypeNames[s.DataType]
		if ok {
			t.Errorf("Expected %s (%d) type but got %s (%d)",
				DataTypeNames[DataTypeString], DataTypeString, dataType, s.DataType)
		} else {
			t.Errorf("Expected %s (%d) type but got %d", DataTypeNames[DataTypeString], DataTypeString, s.DataType)
		}
	}

	path := strings.Join(s.DisplayPath, "/")
	if path != YASTTestSelectorDisplayPath {
		t.Errorf("Expected %s display path but got %s", YASTTestSelectorDisplayPath, path)
	}
}

func TestUnmarshalYASTSelectorWithNetwork(t *testing.T) {
	c, v := prepareTestYAST(YASTTestSelectorWithNetworks, YASTSelectorTestAttrs, YASTSelectorTestContentWithNetworks, t)

	s, err := c.unmarshalSelector(v)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	if s == nil {
		t.Fatalf("Expected selector but got nothing")
	}
}

func TestUnmarshalYASTSelectorInvalid(t *testing.T) {
	c, v := prepareTestYAST(YASTTestInvalidSelector, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err := c.unmarshalSelector(v)
	assertError(err, "Expected selector attributes", t)
}

func TestUnmarshalYASTSelectorInvalidType(t *testing.T) {
	c, v := prepareTestYAST(YASTTestSelectorMissingType, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err := c.unmarshalSelector(v)
	assertError(err, "Missing type", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidTypeType, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected type", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidTypeValue, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Unknown value type", t)
}

func TestUnmarshalYASTSelectorInvalidPath(t *testing.T) {
	c, v := prepareTestYAST(YASTTestSelectorMissingPath, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err := c.unmarshalSelector(v)
	assertError(err, "Missing selector path", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidPathType, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected selector path", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidPathElementType, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected string, value, attribute or selector", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidPathElementMap, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected only one entry in value or attribute map", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidPathElementSpecType, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected specificator", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidPathElementSpec, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected value, attribute or selector specificator", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidPathElementValue, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected value", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidPathElementValueType, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected only string", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidPathElementAttr, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected attribute ID", t)

	c, v = prepareTestYAST(YASTTestSelectorUnknownPathElementAttr, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Unknown attribute ID", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidPathElementAttrType, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected only string or domain", t)
}

func TestUnmarshalYASTSelectorInvalidContent(t *testing.T) {
	c, v := prepareTestYAST(YASTTestSelectorInvalidContent, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err := c.unmarshalSelector(v)
	assertError(err, "Expected selector content", t)

	c, v = prepareTestYAST(YASTTestSelectorUnknownContent, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "No content", t)
}

func TestUnmarshalYASTSelectorDuplicate(t *testing.T) {
	c, v := prepareTestYAST(YASTTestSelector, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	first, err := c.unmarshalSelector(v)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	if first == nil {
		t.Fatal("Expected selector but got nothing")
	}

	second, err := c.unmarshalSelector(v)
	if err != nil {
		t.Fatalf("Expected no errors but got:\n%#v\n\n%s\n", err, err)
	}

	if second == nil {
		t.Fatal("Expected selector but got nothing")
	}

	if first != second {
		t.Errorf("Expected the same selector as a result of parsing the same data got different (%p != %p)", first, second)
	}
}

func TestUnmarshalYASTSelectorInvalidSubselector(t *testing.T) {
	c, v := prepareTestYAST(YASTTestSelectorInvalidSubselector, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err := c.unmarshalSelector(v)
	assertError(err, "Missing type", t)

	c, v = prepareTestYAST(YASTTestSelectorInvalidSubselectorType, YASTSelectorTestAttrs, YASTSelectorTestContent, t)
	_, err = c.unmarshalSelector(v)
	assertError(err, "Expected only string or domain", t)
}
