package pdp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type missingSelectorValue struct {
	err error
}

type SelectorType struct {
	DataType int
	Path     []AttributeDesignatorType
	Content  interface{}

	ContentName string
	DisplayPath []string
}

func (s SelectorType) getResultType() int {
	return s.DataType
}

func castToUndefinedSelectorType(v interface{}, path string) (AttributeValueType, error) {
	return AttributeValueType{}, fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeUndefined])
}

func castToStringSelectorType(v interface{}, path string) (AttributeValueType, error) {
	s, ok := v.(string)
	if !ok {
		return AttributeValueType{},
			fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeString])
	}

	return AttributeValueType{DataTypeString, s}, nil
}

func castToBooleanSelectorType(v interface{}, path string) (AttributeValueType, error) {
	b, ok := v.(bool)
	if !ok {
		return AttributeValueType{},
			fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeBoolean])
	}

	return AttributeValueType{DataTypeBoolean, b}, nil
}

func castToAddressSelectorType(v interface{}, path string) (AttributeValueType, error) {
	s, ok := v.(string)
	if !ok {
		return AttributeValueType{},
			fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeAddress])
	}

	a := net.ParseIP(s)
	if a == nil {
		return AttributeValueType{},
			fmt.Errorf("Can't cast %#v at /%s to %s", s, path, DataTypeNames[DataTypeAddress])
	}

	return AttributeValueType{DataTypeAddress, a}, nil
}

func castToNetworkSelectorType(v interface{}, path string) (AttributeValueType, error) {
	s, ok := v.(string)
	if !ok {
		return AttributeValueType{},
			fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeNetwork])
	}

	_, n, err := net.ParseCIDR(s)
	if err != nil {
		return AttributeValueType{},
			fmt.Errorf("Can't cast %#v at /%s to %s: %v", s, path, DataTypeNames[DataTypeNetwork], err)
	}

	return AttributeValueType{DataTypeNetwork, *n}, nil
}

func castToDomainSelectorType(v interface{}, path string) (AttributeValueType, error) {
	s, ok := v.(string)
	if !ok {
		return AttributeValueType{},
			fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeDomain])
	}

	d, err := AdjustDomainName(s)
	if err != nil {
		return AttributeValueType{},
			fmt.Errorf("Can't cast %T at /%s to %s: %v", s, path, DataTypeNames[DataTypeDomain], err)
	}

	return AttributeValueType{DataTypeDomain, d}, nil
}

func castItemToSetOfStringsSelectorType(item interface{}, i int, path string) (string, error) {
	s, ok := item.(string)
	if !ok {
		return "", fmt.Errorf("Can't cast %d item of type %T at /%s to %s",
			i, item, path, DataTypeNames[DataTypeString])
	}

	return s, nil
}

func castArrayToSetOfStringsSelectorType(v []interface{}, path string) (map[string]bool, error) {
	strs := make(map[string]bool)

	for i, item := range v {
		s, err := castItemToSetOfStringsSelectorType(item, i, path)
		if err != nil {
			return nil, err
		}

		strs[s] = true
	}

	return strs, nil
}

func castMapToSetOfStringsSelectorType(m map[string]interface{}, path string) (map[string]bool, error) {
	strs := make(map[string]bool)

	for k := range m {
		strs[k] = true
	}

	return strs, nil
}

func castToSetOfStringsSelectorType(v interface{}, path string) (AttributeValueType, error) {
	switch items := v.(type) {
	case []interface{}:
		strs, err := castArrayToSetOfStringsSelectorType(items, path)
		if err != nil {
			return AttributeValueType{}, err
		}

		return AttributeValueType{DataTypeSetOfStrings, strs}, nil

	case map[string]interface{}:
		strs, err := castMapToSetOfStringsSelectorType(items, path)
		if err != nil {
			return AttributeValueType{}, err
		}

		return AttributeValueType{DataTypeSetOfStrings, strs}, nil
	}

	return AttributeValueType{},
		fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeSetOfStrings])
}

func castItemToSetOfNetworksSelectorType(item interface{}, i int, path string) (net.IPNet, error) {
	s, ok := item.(string)
	if !ok {
		return net.IPNet{}, fmt.Errorf("Can't cast %d item of type %T at /%s to %s",
			i, item, path, DataTypeNames[DataTypeNetwork])
	}

	_, n, err := net.ParseCIDR(s)
	if err != nil {
		return net.IPNet{}, fmt.Errorf("Can't cast %d item %#v at /%s to %s: %v",
			i, s, path, DataTypeNames[DataTypeNetwork], err)
	}

	return *n, nil
}

func castArrayToSetOfNetworksSelectorType(v []interface{}, path string) ([]net.IPNet, error) {
	nets := make([]net.IPNet, 0)

	for i, item := range v {
		n, err := castItemToSetOfNetworksSelectorType(item, i, path)
		if err != nil {
			return nil, err
		}

		nets = append(nets, n)
	}

	return nets, nil
}

func castKeyToSetOfNetworksSelectorType(k string, path string) (net.IPNet, error) {
	_, n, err := net.ParseCIDR(k)
	if err != nil {
		return net.IPNet{}, fmt.Errorf("Can't cast %#v at /%s to %s: %v",
			k, path, DataTypeNames[DataTypeNetwork], err)
	}

	return *n, nil
}

func castMapToSetOfNetworksSelectorType(m map[string]interface{}, path string) ([]net.IPNet, error) {
	nets := make([]net.IPNet, 0)

	for k := range m {
		n, err := castKeyToSetOfNetworksSelectorType(k, path)
		if err != nil {
			return nil, err
		}

		nets = append(nets, n)
	}

	return nets, nil
}

func castToSetOfNetworksSelectorType(v interface{}, path string) (AttributeValueType, error) {
	switch items := v.(type) {
	case []interface{}:
		nets, err := castArrayToSetOfNetworksSelectorType(items, path)
		if err != nil {
			return AttributeValueType{}, err
		}

		return AttributeValueType{DataTypeSetOfNetworks, nets}, nil

	case map[string]interface{}:
		nets, err := castMapToSetOfNetworksSelectorType(items, path)
		if err != nil {
			return AttributeValueType{}, err
		}

		return AttributeValueType{DataTypeSetOfNetworks, nets}, nil
	}

	return AttributeValueType{},
		fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeSetOfNetworks])
}

func castArrayToSetOfDomainsSelectorType(v []interface{}, path string) (SetOfSubdomains, error) {
	set := SetOfSubdomains{false, nil, make(map[string]*SetOfSubdomains)}

	for i, item := range v {
		s, ok := item.(string)
		if !ok {
			return set, fmt.Errorf("Can't cast %d item of type %T at /%s to %s",
				i, item, path, DataTypeNames[DataTypeDomain])
		}

		d, err := AdjustDomainName(s)
		if err != nil {
			return set, fmt.Errorf("Can't cast %d item %#v at /%s to %s: %v",
				i, item, path, DataTypeNames[DataTypeDomain], err)
		}

		set.addToSetOfDomains(d, nil)
	}

	return set, nil
}

func castMapToSetOfDomainsSelectorType(m map[string]interface{}, path string) (SetOfSubdomains, error) {
	set := SetOfSubdomains{false, nil, make(map[string]*SetOfSubdomains)}

	for k := range m {
		d, err := AdjustDomainName(k)
		if err != nil {
			return set, fmt.Errorf("Can't cast %#v at /%s to %s: %v",
				k, path, DataTypeNames[DataTypeDomain], err)
		}

		set.addToSetOfDomains(d, nil)
	}

	return set, nil
}

func castToSetOfDomainsSelectorType(v interface{}, path string) (AttributeValueType, error) {
	switch items := v.(type) {
	case []interface{}:
		d, err := castArrayToSetOfDomainsSelectorType(items, path)
		if err != nil {
			return AttributeValueType{}, err
		}

		return AttributeValueType{DataTypeSetOfDomains, d}, nil

	case map[string]interface{}:
		d, err := castMapToSetOfDomainsSelectorType(items, path)
		if err != nil {
			return AttributeValueType{}, err
		}

		return AttributeValueType{DataTypeSetOfDomains, d}, nil
	}

	return AttributeValueType{},
		fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeSetOfDomains])
}

func castToSelectorType(v interface{}, t int, path string) (AttributeValueType, error) {
	switch t {
	case DataTypeUndefined:
		return castToUndefinedSelectorType(v, path)

	case DataTypeString:
		return castToStringSelectorType(v, path)

	case DataTypeBoolean:
		return castToBooleanSelectorType(v, path)

	case DataTypeAddress:
		return castToAddressSelectorType(v, path)

	case DataTypeNetwork:
		return castToNetworkSelectorType(v, path)

	case DataTypeDomain:
		return castToDomainSelectorType(v, path)

	case DataTypeSetOfStrings:
		return castToSetOfStringsSelectorType(v, path)

	case DataTypeSetOfNetworks:
		return castToSetOfNetworksSelectorType(v, path)

	case DataTypeSetOfDomains:
		return castToSetOfDomainsSelectorType(v, path)
	}

	return AttributeValueType{},
		fmt.Errorf("Cast to %s type hasn't been implemented yet (%s)", DataTypeNames[t], path)
}

func castMissingSelectorValue(t int, err error) (AttributeValueType, error) {
	switch t {
	case DataTypeSetOfStrings:
		return AttributeValueType{DataTypeSetOfStrings, make(map[string]bool)}, nil

	case DataTypeSetOfNetworks:
		return AttributeValueType{DataTypeSetOfNetworks, make([]net.IPNet, 0)}, nil

	case DataTypeSetOfDomains:
		return AttributeValueType{DataTypeSetOfDomains, SetOfSubdomains{false, nil, make(map[string]*SetOfSubdomains)}}, nil
	}

	return AttributeValueType{}, err
}

func dispatchContentByType(c interface{}, path []string) (interface{}, AttributeValueType, bool, error) {
	switch v := c.(type) {
	case *SetOfSubdomains, map[string]interface{}:
		return v, AttributeValueType{}, false, nil

	case AttributeValueType:
		return nil, v, false, nil

	case missingSelectorValue:
		return nil, AttributeValueType{}, true, v.err

	case error:
		return nil, AttributeValueType{}, false, v
	}

	return nil, AttributeValueType{}, false,
		fmt.Errorf("Expected map, domain set, value, missing value or error at /%s but got %T (%#v)", strings.Join(path, "/"), c, c)
}

func (s SelectorType) calculate(ctx *Context) (AttributeValueType, error) {
	path := []string{}

	m, v, miss, err := dispatchContentByType(s.Content, path)
	if miss {
		return castMissingSelectorValue(s.DataType, err)
	}

	if err != nil {
		return AttributeValueType{}, err
	}

	for i, item := range s.Path {
		if m == nil {
			return AttributeValueType{},
				fmt.Errorf("Expected map, domain set, missing value or error at /%s but got attribute value (%#v)",
					strings.Join(path, "/"), v)
		}

		path = append(path, s.DisplayPath[i])

		v, err = item.calculate(ctx)
		if err != nil {
			return AttributeValueType{},
				fmt.Errorf("Error on calculating %s at /%s: %v", item.Attribute.ID, strings.Join(path, "/"), err)
		}

		var idx string
		switch v.DataType {
		default:
			return AttributeValueType{},
				fmt.Errorf("Expected string or domain as %s at /%s but got %s",
					item.Attribute.ID, strings.Join(path, "/"), DataTypeNames[v.DataType])

		case DataTypeString:
			idx, err = ExtractStringValue(v, fmt.Sprintf("%s at /%s", item.Attribute.ID, strings.Join(path, "/")))
			if err != nil {
				return AttributeValueType{}, err
			}

		case DataTypeDomain:
			idx, err = ExtractDomainValue(v, fmt.Sprintf("%s at /%s", item.Attribute.ID, strings.Join(path, "/")))
			if err != nil {
				return AttributeValueType{}, err
			}
		}

		path[len(path)-1] = fmt.Sprintf("%s(%s)", path[len(path)-1], idx)

		var (
			c  interface{}
			ok bool
		)

		switch m := m.(type) {
		default:
			return AttributeValueType{},
				fmt.Errorf("Expected map or domain set at /%s but got %T (%#v)", strings.Join(path, "/"), m, m)

		case map[string]interface{}:
			c, ok = m[idx]
			if !ok {
				err := fmt.Errorf("No value at /%s", strings.Join(path, "/"))
				return castMissingSelectorValue(s.DataType, &MissingValueError{err})
			}

		case *SetOfSubdomains:
			c, ok = m.Get(idx)
			if !ok {
				err := fmt.Errorf("No value at /%s", strings.Join(path, "/"))
				return castMissingSelectorValue(s.DataType, &MissingValueError{err})
			}
		}

		m, v, miss, err = dispatchContentByType(c, path)
		if miss {
			return castMissingSelectorValue(s.DataType, err)
		}

		if err != nil {
			return AttributeValueType{}, err
		}
	}

	return v, nil
}

func duckToSelectorContentbyAttribute(a AttributeDesignatorType, c interface{}, rawPath []ExpressionType, t int, reprPath []string) interface{} {
	m, ok := c.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Expected map at /%s but got %T", strings.Join(reprPath, "/"), c)
	}

	if a.getResultType() == DataTypeDomain {
		s := &SetOfSubdomains{false, nil, make(map[string]*SetOfSubdomains)}
		for k, v := range m {
			subReprPath := append(reprPath, fmt.Sprintf("%s(%q)", DataTypeNames[DataTypeDomain], k))
			s.addToSetOfDomains(k, duckToSelectorContent(v, rawPath, t, subReprPath))
		}

		return s
	}

	r := make(map[string]interface{})
	for k, v := range m {
		subReprPath := append(reprPath, fmt.Sprintf("%q", k))
		r[k] = duckToSelectorContent(v, rawPath, t, subReprPath)
	}

	return r
}

func fetchFromSelectorContentArray(c []interface{}, s string, reprPath []string) interface{} {
	i, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("Expected integer as path component at /%s but got %#v",
			strings.Join(reprPath, "/"), s)
	}

	if i < 0 || i >= len(c) {
		return fmt.Errorf("Index %d out of bounds for array of length %d at /%s",
			i, len(c), strings.Join(reprPath, "/"))
	}

	return c[i]
}

func fetchFromSelectorContentMap(c map[string]interface{}, s string, reprPath []string) interface{} {
	v, ok := c[s]
	if !ok {
		err := fmt.Errorf("Missing value for key %s at /%s", s, strings.Join(reprPath, "/"))
		return missingSelectorValue{&MissingValueError{err}}
	}

	return v
}

func duckToSelectorContent(rawCtx interface{}, rawPath []ExpressionType, t int, reprPath []string) interface{} {
	c := rawCtx
	var err error

	for i, item := range rawPath {
		switch v := item.(type) {
		case AttributeDesignatorType:
			return duckToSelectorContentbyAttribute(v, c, rawPath[i+1:], t, reprPath)

		case AttributeValueType:
			s := v.Value.(string)

			a, ok := c.([]interface{})
			if ok {
				c = fetchFromSelectorContentArray(a, s, reprPath)
				err, ok := c.(error)
				if ok {
					return err
				}

				reprPath = append(reprPath, fmt.Sprintf("%q", s))
				continue
			}

			m, ok := c.(map[string]interface{})
			if ok {
				c = fetchFromSelectorContentMap(m, s, reprPath)
				miss, ok := c.(missingSelectorValue)
				if ok {
					return miss
				}

				reprPath = append(reprPath, fmt.Sprintf("%q", s))
				continue
			}

			return fmt.Errorf("Expected object or array at /%s but got %T", strings.Join(reprPath, "/"), c)
		}
	}

	v, err := castToSelectorType(c, t, strings.Join(reprPath, "/"))
	if err != nil {
		return err
	}

	return v
}

func prepareSelectorPath(raw []ExpressionType) ([]AttributeDesignatorType, []string) {
	path := []AttributeDesignatorType{}
	displayPath := []string{}
	displayPathItem := []string{}

	for _, item := range raw {
		switch v := item.(type) {
		case AttributeValueType:
			displayPathItem = append(displayPathItem, fmt.Sprintf("%q", v.Value.(string)))

		case AttributeDesignatorType:
			displayPathItem = append(displayPathItem, fmt.Sprintf("%s(%q)", yastTagAttribute, v.Attribute.ID))
			displayPath = append(displayPath, strings.Join(displayPathItem, "/"))
			displayPathItem = []string{}

			path = append(path, v)
		}
	}

	return path, displayPath
}

func prepareSelectorContent(rawCtx interface{}, rawPath []ExpressionType, t int) (interface{}, []AttributeDesignatorType, []string) {
	ctx := duckToSelectorContent(rawCtx, rawPath, t, []string{})
	path, displayPath := prepareSelectorPath(rawPath)

	return ctx, path, displayPath
}
