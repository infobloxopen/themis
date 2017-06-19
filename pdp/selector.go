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
	Path     []ExpressionType
	Content  interface{}

	ContentName string
	DisplayPath []string
}

func (s SelectorType) describe() string {
	return fmt.Sprintf("%s(%q:%s)", yastTagSelector, s.ContentName, strings.Join(s.DisplayPath, "/"))
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

func castItemToCollectionOfStringsSelectorType(item interface{}, i int, path string) (string, error) {
	s, ok := item.(string)
	if !ok {
		return "", fmt.Errorf("Can't cast %d item of type %T at /%s to %s",
			i, item, path, DataTypeNames[DataTypeString])
	}

	return s, nil
}

func castArrayToSetOfStringsSelectorType(v []interface{}, path string) (map[string]int, error) {
	strs := make(map[string]int)

	for i, item := range v {
		s, err := castItemToCollectionOfStringsSelectorType(item, i, path)
		if err != nil {
			return nil, err
		}

		strs[s] = i
	}

	return strs, nil
}

func castMapToSetOfStringsSelectorType(m map[string]interface{}, path string) (map[string]int, error) {
	strs := make(map[string]int)

	i := 0
	for k := range m {
		strs[k] = i
		i++
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

func castItemToSetOfNetworksSelectorType(item interface{}, i int, path string) (*net.IPNet, error) {
	s, ok := item.(string)
	if !ok {
		return nil, fmt.Errorf("Can't cast %d item of type %T at /%s to %s",
			i, item, path, DataTypeNames[DataTypeNetwork])
	}

	n, err := MakeNetwork(s)
	if err != nil {
		return nil, fmt.Errorf("Can't cast %d item %#v at /%s to %s: %v",
			i, item, path, DataTypeNames[DataTypeNetwork], err)
	}

	return n, nil
}

func castArrayToSetOfNetworksSelectorType(v []interface{}, path string) (*SetOfNetworks, error) {
	nets := NewSetOfNetworks()

	for i, item := range v {
		n, err := castItemToSetOfNetworksSelectorType(item, i, path)
		if err != nil {
			return nil, err
		}

		nets.addToSetOfNetworks(n, true)
	}

	return nets, nil
}

func castKeyToSetOfNetworksSelectorType(k string, path string) (*net.IPNet, error) {
	n, err := MakeNetwork(k)
	if err != nil {
		return nil, fmt.Errorf("Can't cast %#v at /%s to %s: %v",
			k, path, DataTypeNames[DataTypeNetwork], err)
	}

	return n, nil
}

func castMapToSetOfNetworksSelectorType(m map[string]interface{}, path string) (*SetOfNetworks, error) {
	nets := NewSetOfNetworks()

	for k := range m {
		n, err := castKeyToSetOfNetworksSelectorType(k, path)
		if err != nil {
			return nil, err
		}

		nets.addToSetOfNetworks(n, true)
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

func castArrayToSetOfDomainsSelectorType(v []interface{}, path string) (*SetOfSubdomains, error) {
	set := NewSetOfSubdomains()

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

		set.insert(d, i)
	}

	return set, nil
}

func castMapToSetOfDomainsSelectorType(m map[string]interface{}, path string) (*SetOfSubdomains, error) {
	set := NewSetOfSubdomains()

	for k := range m {
		d, err := AdjustDomainName(k)
		if err != nil {
			return set, fmt.Errorf("Can't cast %#v at /%s to %s: %v",
				k, path, DataTypeNames[DataTypeDomain], err)
		}

		set.insert(d, nil)
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

func castArrayToListOfStringsSelectorType(v []interface{}, path string) ([]string, error) {
	strs := []string{}

	for i, item := range v {
		s, err := castItemToCollectionOfStringsSelectorType(item, i, path)
		if err != nil {
			return nil, err
		}

		strs = append(strs, s)
	}

	return strs, nil
}

func castToListOfStringsSelectorType(v interface{}, path string) (AttributeValueType, error) {
	switch items := v.(type) {
	case []interface{}:
		strs, err := castArrayToListOfStringsSelectorType(items, path)
		if err != nil {
			return AttributeValueType{}, err
		}

		return AttributeValueType{DataTypeListOfStrings, strs}, nil
	}

	return AttributeValueType{},
		fmt.Errorf("Can't cast %T at /%s to %s", v, path, DataTypeNames[DataTypeListOfStrings])
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

	case DataTypeListOfStrings:
		return castToListOfStringsSelectorType(v, path)
	}

	return AttributeValueType{},
		fmt.Errorf("Cast to %s type hasn't been implemented yet (%s)", DataTypeNames[t], path)
}

func castMissingSelectorValue(t int, err error) (AttributeValueType, error) {
	switch t {
	case DataTypeSetOfStrings:
		return AttributeValueType{DataTypeSetOfStrings, make(map[string]int)}, nil

	case DataTypeSetOfNetworks:
		return AttributeValueType{DataTypeSetOfNetworks, NewSetOfNetworks()}, nil

	case DataTypeSetOfDomains:
		return AttributeValueType{DataTypeSetOfDomains, NewSetOfSubdomains()}, nil

	case DataTypeListOfStrings:
		return AttributeValueType{DataTypeListOfStrings, []string{}}, nil
	}

	return AttributeValueType{}, err
}

func dispatchContentByType(c interface{}, path []string) (interface{}, AttributeValueType, bool, error) {
	switch v := c.(type) {
	case *SetOfSubdomains, *SetOfNetworks, map[string]interface{}:
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
				fmt.Errorf("Error on calculating %s at /%s: %v", item.describe(), strings.Join(path, "/"), err)
		}

		var idx interface{}
		switch v.DataType {
		default:
			return AttributeValueType{},
				fmt.Errorf("Expected string, domain, address or network as %s at /%s but got %s",
					item.describe(), strings.Join(path, "/"), DataTypeNames[v.DataType])

		case DataTypeString:
			idx, err = ExtractStringValue(v, fmt.Sprintf("%s at /%s", item.describe(), strings.Join(path, "/")))

		case DataTypeDomain:
			idx, err = ExtractDomainValue(v, fmt.Sprintf("%s at /%s", item.describe(), strings.Join(path, "/")))

		case DataTypeAddress:
			idx, err = ExtractAddressValue(v, fmt.Sprintf("%s at /%s", item.describe(), strings.Join(path, "/")))

		case DataTypeNetwork:
			idx, err = ExtractNetworkValue(v, fmt.Sprintf("%s at /%s", item.describe(), strings.Join(path, "/")))
		}

		if err != nil {
			return AttributeValueType{}, err
		}

		var (
			c  interface{}
			ok bool
		)

		switch m := m.(type) {
		default:
			return AttributeValueType{},
				fmt.Errorf("Expected map, domain set or network set at /%s but got %T (%#v)", strings.Join(path, "/"), m, m)

		case map[string]interface{}:
			path[len(path)-1] = fmt.Sprintf("%s(%s)", path[len(path)-1], idx)
			c, ok = m[idx.(string)]
		case *SetOfSubdomains:
			path[len(path)-1] = fmt.Sprintf("%s(%s)", path[len(path)-1], idx)
			c, ok = m.Get(idx.(string))

		case *SetOfNetworks:
			switch idx := idx.(type) {
			case net.IP:
				path[len(path)-1] = fmt.Sprintf("%s(%s)", path[len(path)-1], idx.String())
				c = m.GetByAddr(idx)
				ok = c != nil

			case net.IPNet:
				path[len(path)-1] = fmt.Sprintf("%s(%s)", path[len(path)-1], idx.String())
				c = m.GetByNet(&idx)
				ok = c != nil
			}
		}

		if !ok {
			err := fmt.Errorf("No value at /%s", strings.Join(path, "/"))
			return castMissingSelectorValue(s.DataType, &MissingValueError{err})
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

func duckToStringSetSelectorContent(m map[string]interface{}, rawPath []ExpressionType, t int, reprPath []string) interface{} {
	s := make(map[string]interface{})
	for k, v := range m {
		subReprPath := append(reprPath, fmt.Sprintf("%q", k))
		s[k] = duckToSelectorContent(v, rawPath, t, subReprPath)
	}

	return s
}

func duckToDomainsSetSelectorContent(m map[string]interface{}, rawPath []ExpressionType, t int, reprPath []string) interface{} {
	s := NewSetOfSubdomains()
	for k, v := range m {
		subReprPath := append(reprPath, fmt.Sprintf("%s(%q)", DataTypeNames[DataTypeDomain], k))
		d, err := AdjustDomainName(k)
		if err != nil {
			return fmt.Errorf("Can't cast %#v at /%s to %s: %v",
				k, strings.Join(reprPath, "/"), DataTypeNames[DataTypeDomain], err)
		}

		s.insert(d, duckToSelectorContent(v, rawPath, t, subReprPath))
	}

	return s
}

func duckToNetworkSetSelectorContent(m map[string]interface{}, rawPath []ExpressionType, t int, reprPath []string) interface{} {
	s := NewSetOfNetworks()
	for k, v := range m {
		subReprPath := append(reprPath, fmt.Sprintf("%s(%q)", DataTypeNames[DataTypeNetwork], k))
		n, err := MakeNetwork(k)
		if err != nil {
			return fmt.Errorf("Can't cast %#v at /%s to %s: %v",
				k, strings.Join(reprPath, "/"), DataTypeNames[DataTypeNetwork], err)
		}

		s.addToSetOfNetworks(n, duckToSelectorContent(v, rawPath, t, subReprPath))
	}

	return s
}

func duckToSelectorContentbyExpression(a ExpressionType, c interface{}, rawPath []ExpressionType, t int, reprPath []string) interface{} {
	m, ok := c.(map[string]interface{})
	if !ok {
		return fmt.Errorf("Expected map at /%s but got %T", strings.Join(reprPath, "/"), c)
	}

	switch a.getResultType() {
	case DataTypeDomain:
		return duckToDomainsSetSelectorContent(m, rawPath, t, reprPath)

	case DataTypeAddress, DataTypeNetwork:
		return duckToNetworkSetSelectorContent(m, rawPath, t, reprPath)
	}

	return duckToStringSetSelectorContent(m, rawPath, t, reprPath)

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
		default:
			return fmt.Errorf("Expected value or attribute after /%s but got %T", strings.Join(reprPath, "/"), v)

		case *SelectorType:
			return duckToSelectorContentbyExpression(v, c, rawPath[i+1:], t, reprPath)

		case AttributeDesignatorType:
			return duckToSelectorContentbyExpression(v, c, rawPath[i+1:], t, reprPath)

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

func prepareSelectorPath(raw []ExpressionType) ([]ExpressionType, []string) {
	path := []ExpressionType{}
	displayPath := []string{}
	displayPathItem := []string{}

	for _, item := range raw {
		displayPathItem = append(displayPathItem, item.describe())

		switch v := item.(type) {
		case AttributeValueType:
			break

		case AttributeDesignatorType:
			displayPath = append(displayPath, strings.Join(displayPathItem, "/"))
			displayPathItem = []string{}

			path = append(path, v)

		case *SelectorType:
			displayPath = append(displayPath, strings.Join(displayPathItem, "/"))
			displayPathItem = []string{}

			path = append(path, v)
		}
	}

	if len(displayPathItem) > 0 {
		displayPath = append(displayPath, strings.Join(displayPathItem, "/"))
	}

	return path, displayPath
}

func prepareSelectorContent(rawCtx interface{}, rawPath []ExpressionType, t int) (interface{}, []ExpressionType, []string) {
	ctx := duckToSelectorContent(rawCtx, rawPath, t, []string{})
	path, displayPath := prepareSelectorPath(rawPath)

	return ctx, path, displayPath
}
