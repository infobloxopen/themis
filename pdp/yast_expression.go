package pdp

import (
	"net"
	"strings"
)

type anyArgumentsFunctionType func(args []ExpressionType) ExpressionType
type argumentChecker func(args []ExpressionType) anyArgumentsFunctionType

var expressionArgumentCheckers map[string][]argumentChecker = map[string][]argumentChecker{
	yastExpressionEqual: {checkerFunctionStringEqual},
	yastExpressionContains: {
		checkerFunctionStringContains,
		checkerFunctionNetworkContainsAddress,
		checkerFunctionSetOfStringContains,
		checkerFunctionSetOfNetworksContainsAddress,
		checkerFunctionSetOfDomainsContains},
	yastExpressionNot: {checkerFunctionBooleanNot},
	yastExpressionOr:  {checkerFunctionBooleanOr},
	yastExpressionAnd: {checkerFunctionBooleanAnd}}

func (ctx YastCtx) unmarshalStringValue(v interface{}) (*AttributeValueType, error) {
	s, err := ctx.validateString(v, "value of string type")
	if err != nil {
		return nil, err
	}

	return &AttributeValueType{DataTypeString, s}, nil
}

func (ctx YastCtx) unmarshalAddressValue(v interface{}) (*AttributeValueType, error) {
	s, err := ctx.validateString(v, "value of address type")
	if err != nil {
		return nil, err
	}

	a := net.ParseIP(s)
	if a == nil {
		return nil, ctx.errorf("Expected value of address type but got %#v", s)
	}

	return &AttributeValueType{DataTypeAddress, a}, nil
}

func (ctx YastCtx) unmarshalNetworkValue(v interface{}) (*AttributeValueType, error) {
	s, err := ctx.validateString(v, "value of network type")
	if err != nil {
		return nil, err
	}

	_, n, err := net.ParseCIDR(s)
	if err != nil {
		return nil, ctx.errorf("Expected value of network type but got %#v (%v)", s, err)
	}

	return &AttributeValueType{DataTypeNetwork, *n}, nil
}

func (ctx YastCtx) unmarshalDomainValue(v interface{}) (*AttributeValueType, error) {
	s, err := ctx.validateString(v, "value of domain type")
	if err != nil {
		return nil, err
	}

	d, err := AdjustDomainName(s)
	if err != nil {
		return nil, ctx.errorf("Expected value of domain type but got %#v (%v)", s, err)
	}

	return &AttributeValueType{DataTypeDomain, d}, nil
}

func (ctx *YastCtx) unmarshalSetOfStringsValueItem(v interface{}, i int, set map[string]int) error {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	s, err := ctx.validateString(v, "element of set of strings")
	if err != nil {
		return err
	}

	set[s] = i
	return nil
}

func (ctx *YastCtx) unmarshalSetOfStringsImmediateValue(v interface{}) (*AttributeValueType, error) {
	items, err := ctx.validateList(v, "")
	if err != nil {
		return nil, nil
	}

	set := make(map[string]int)
	for i, item := range items {
		err = ctx.unmarshalSetOfStringsValueItem(item, i, set)
		if err != nil {
			return nil, err
		}
	}

	return &AttributeValueType{DataTypeSetOfStrings, set}, nil
}

func (ctx *YastCtx) unmarshalSetOfStringsValueFromContent(v interface{}) (*AttributeValueType, error) {
	set, err := ctx.extractContentByItem(v)
	if err != nil || set == nil {
		return nil, err
	}

	return ctx.unmarshalSetOfStringsImmediateValue(set)
}

func (ctx *YastCtx) unmarshalSetOfStringsValue(v interface{}) (*AttributeValueType, error) {
	val, err := ctx.unmarshalSetOfStringsImmediateValue(v)
	if err != nil {
		return nil, err
	}

	if val != nil {
		return val, nil
	}

	val, err = ctx.unmarshalSetOfStringsValueFromContent(v)
	if err != nil {
		return nil, err
	}

	if val != nil {
		return val, nil
	}

	return nil, ctx.errorf("Expected value of set of strings type or content id but got %v", v)
}

func (ctx *YastCtx) unmarshalSetOfNetworksValueItem(v interface{}, i int, set *SetOfNetworks) error {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	s, err := ctx.validateString(v, "element of set of networks")
	if err != nil {
		return err
	}

	n, err := MakeNetwork(s)
	if err != nil {
		if err == ErrorIPv6NotImplemented {
			return ctx.errorf("Got IPv6 value %#v but it isn't supported", s)
		}

		return ctx.errorf("Expected value of network type but got %#v (%v)", s, err)
	}

	set.addToSetOfNetworks(n, true)
	return nil
}

func (ctx *YastCtx) unmarshalSetOfNetworksImmediateValue(v interface{}) (*AttributeValueType, error) {
	items, err := ctx.validateList(v, "")
	if err != nil {
		return nil, nil
	}

	set := NewSetOfNetworks()
	for i, item := range items {
		err = ctx.unmarshalSetOfNetworksValueItem(item, i, set)
		if err != nil {
			return nil, err
		}
	}

	return &AttributeValueType{DataTypeSetOfNetworks, set}, nil
}

func (ctx *YastCtx) unmarshalSetOfNetworksValueFromContent(v interface{}) (*AttributeValueType, error) {
	set, err := ctx.extractContentByItem(v)
	if err != nil || set == nil {
		return nil, err
	}

	return ctx.unmarshalSetOfNetworksImmediateValue(set)
}

func (ctx *YastCtx) unmarshalSetOfNetworksValue(v interface{}) (*AttributeValueType, error) {
	val, err := ctx.unmarshalSetOfNetworksImmediateValue(v)
	if err != nil {
		return nil, err
	}

	if val != nil {
		return val, nil
	}

	val, err = ctx.unmarshalSetOfNetworksValueFromContent(v)
	if err != nil {
		return nil, err
	}

	if val != nil {
		return val, nil
	}

	return nil, ctx.errorf("Expected value of set of networks type or content id but got %v", v)
}

func (ctx *YastCtx) unmarshalSetOfDomainsValueItem(v interface{}, i int, set *SetOfSubdomains) error {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	s, err := ctx.validateString(v, "element of set of domains")
	if err != nil {
		return err
	}

	d, err := AdjustDomainName(s)
	if err != nil {
		return ctx.errorf("Expected value of domain type but got %#v (%v)", s, err)
	}

	set.insert(d, i)

	return nil
}

func (ctx *YastCtx) unmarshalSetOfDomainsImmediateValue(v interface{}) (*AttributeValueType, error) {
	items, err := ctx.validateList(v, "")
	if err != nil {
		return nil, nil
	}

	set := NewSetOfSubdomains()
	for i, item := range items {
		err = ctx.unmarshalSetOfDomainsValueItem(item, i, set)
		if err != nil {
			return nil, err
		}
	}

	return &AttributeValueType{DataTypeSetOfDomains, set}, nil
}

func (ctx *YastCtx) unmarshalSetOfDomainsValueFromContent(v interface{}) (*AttributeValueType, error) {
	set, err := ctx.extractContentByItem(v)
	if err != nil || set == nil {
		return nil, err
	}

	return ctx.unmarshalSetOfDomainsImmediateValue(set)
}

func (ctx *YastCtx) unmarshalSetOfDomainsValue(v interface{}) (*AttributeValueType, error) {
	val, err := ctx.unmarshalSetOfDomainsImmediateValue(v)
	if err != nil {
		return nil, err
	}

	if val != nil {
		return val, nil
	}

	val, err = ctx.unmarshalSetOfDomainsValueFromContent(v)
	if err != nil {
		return nil, err
	}

	if val != nil {
		return val, nil
	}

	return nil, ctx.errorf("Expected value of set of domains type or content id but got %v", v)
}

func (ctx *YastCtx) unmarshalListOfStringsValueItem(v interface{}, i int, list []string) ([]string, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	s, err := ctx.validateString(v, "element of set of strings")
	if err != nil {
		return list, err
	}

	return append(list, s), nil
}

func (ctx *YastCtx) unmarshalListOfStringsImmediateValue(v interface{}) (*AttributeValueType, error) {
	items, err := ctx.validateList(v, "")
	if err != nil {
		return nil, nil
	}

	list := []string{}
	for i, item := range items {
		list, err = ctx.unmarshalListOfStringsValueItem(item, i, list)
		if err != nil {
			return nil, err
		}
	}

	return &AttributeValueType{DataTypeListOfStrings, list}, nil
}

func (ctx *YastCtx) unmarshalListOfStringsValueFromContent(v interface{}) (*AttributeValueType, error) {
	set, err := ctx.extractContentByItem(v)
	if err != nil || set == nil {
		return nil, err
	}

	return ctx.unmarshalListOfStringsImmediateValue(set)
}

func (ctx *YastCtx) unmarshalListOfStringsValue(v interface{}) (*AttributeValueType, error) {
	val, err := ctx.unmarshalListOfStringsImmediateValue(v)
	if err != nil {
		return nil, err
	}

	if val != nil {
		return val, nil
	}

	val, err = ctx.unmarshalListOfStringsValueFromContent(v)
	if err != nil {
		return nil, err
	}

	if val != nil {
		return val, nil
	}

	return nil, ctx.errorf("Expected value of list of strings type or content id but got %v", v)
}

func (ctx *YastCtx) unmarshalValueByType(t int, v interface{}) (*AttributeValueType, error) {
	if t == DataTypeUndefined {
		return nil, ctx.errorf("Not allowed type %#v", DataTypeNames[t])
	}

	switch t {
	case DataTypeString:
		return ctx.unmarshalStringValue(v)

	case DataTypeAddress:
		return ctx.unmarshalAddressValue(v)

	case DataTypeNetwork:
		return ctx.unmarshalNetworkValue(v)

	case DataTypeDomain:
		return ctx.unmarshalDomainValue(v)

	case DataTypeSetOfStrings:
		return ctx.unmarshalSetOfStringsValue(v)

	case DataTypeSetOfNetworks:
		return ctx.unmarshalSetOfNetworksValue(v)

	case DataTypeSetOfDomains:
		return ctx.unmarshalSetOfDomainsValue(v)

	case DataTypeListOfStrings:
		return ctx.unmarshalListOfStringsValue(v)
	}

	return nil, ctx.errorf("Parsing for type %s hasn't been implemented yet", DataTypeNames[t])
}

func (ctx *YastCtx) unmarshalValue(v interface{}) (AttributeValueType, error) {
	ctx.pushNodeSpec(yastTagValue)
	defer ctx.popNodeSpec()

	m, err := ctx.validateMap(v, "value attributes")
	if err != nil {
		return AttributeValueType{}, err
	}

	strT, err := ctx.extractString(m, yastTagType, "type")
	if err != nil {
		return AttributeValueType{}, err
	}

	t, ok := DataTypeIDs[strings.ToLower(strT)]
	if !ok {
		return AttributeValueType{}, ctx.errorf("Unknown value type %#v", strT)
	}

	c, ok := m[yastTagContent]
	if !ok {
		return AttributeValueType{}, ctx.errorf("No content")
	}

	val, err := ctx.unmarshalValueByType(t, c)
	if err != nil {
		return AttributeValueType{}, err
	}

	return *val, nil
}

func (ctx *YastCtx) unmarshalArgument(v interface{}, i int, exprs []ExpressionType) ([]ExpressionType, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	e, err := ctx.unmarshalExpression(v)
	if err != nil {
		return nil, err
	}

	return append(exprs, e), nil
}

func (ctx *YastCtx) unmarshalArguments(v interface{}) ([]ExpressionType, error) {
	items, err := ctx.validateList(v, "arguments")
	if err != nil {
		return nil, err
	}

	ctx.pushNodeSpec("arguments")
	defer ctx.popNodeSpec()

	exprs := make([]ExpressionType, 0)
	for i, item := range items {
		exprs, err = ctx.unmarshalArgument(item, i, exprs)
		if err != nil {
			return nil, err
		}
	}

	return exprs, nil
}

func (ctx *YastCtx) unmarshalExpression(expr interface{}) (ExpressionType, error) {
	e, err := ctx.validateMap(expr, "expression")
	if err != nil {
		return nil, err
	}

	k, v, err := ctx.getSingleMapPair(e, "expression map")
	if err != nil {
		return nil, err
	}

	s, err := ctx.validateString(k, "specificator or function name")
	if err != nil {
		return nil, err
	}

	if s == yastTagAttribute {
		ID, err := ctx.validateString(v, "attribute ID")
		if err != nil {
			return nil, err
		}

		a, ok := ctx.attrs[ID]
		if !ok {
			return nil, ctx.errorf("Unknown attribute ID %s", ID)
		}

		return AttributeDesignatorType{a}, nil
	}

	if s == yastTagValue {
		return ctx.unmarshalValue(v)
	}

	if s == yastTagSelector {
		return ctx.unmarshalSelector(v)
	}

	checks, ok := expressionArgumentCheckers[s]
	if !ok {
		return nil, ctx.errorf("Expected attribute, immediate value, selector or function but got %s", s)
	}

	ctx.pushNodeSpec("%#v", s)
	defer ctx.popNodeSpec()

	exprs, err := ctx.unmarshalArguments(v)
	if err != nil {
		return nil, err
	}

	for _, c := range checks {
		if m := c(exprs); m != nil {
			return m(exprs), nil
		}
	}

	if len(exprs) > 1 {
		types := make([]string, 0)
		for _, e := range exprs {
			types = append(types, DataTypeNames[e.getResultType()])
		}

		return nil, ctx.errorf("Can't find function %s which takes %d arguments of following types %s",
			s, len(exprs), strings.Join(types, ", "))
	}

	if len(exprs) > 0 {
		return nil, ctx.errorf("Can't find function %s which takes argument of type %s",
			s, DataTypeNames[exprs[0].getResultType()])
	}

	return nil, ctx.errorf("Can't find function %s which takes no arguments", s)
}
