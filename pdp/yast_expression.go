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

func (ctx yastCtx) unmarshalStringValue(v interface{}) (*AttributeValueType, error) {
	s, err := ctx.validateString(v, "value of string type")
	if err != nil {
		return nil, err
	}

	return &AttributeValueType{DataTypeString, s}, nil
}

func (ctx yastCtx) unmarshalAddressValue(v interface{}) (*AttributeValueType, error) {
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

func (ctx yastCtx) unmarshalNetworkValue(v interface{}) (*AttributeValueType, error) {
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

func (ctx yastCtx) unmarshalDomainValue(v interface{}) (*AttributeValueType, error) {
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

func (ctx *yastCtx) unmarshalSetOfStringsValueItem(v interface{}, i int, set *map[string]bool) error {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	s, err := ctx.validateString(v, "element of set of strings")
	if err != nil {
		return err
	}

	(*set)[s] = true
	return nil
}

func (ctx *yastCtx) unmarshalSetOfStringsImmediateValue(v interface{}) (*AttributeValueType, error) {
	items, err := ctx.validateList(v, "")
	if err != nil {
		return nil, nil
	}

	set := make(map[string]bool)
	for i, item := range items {
		err = ctx.unmarshalSetOfStringsValueItem(item, i, &set)
		if err != nil {
			return nil, err
		}
	}

	return &AttributeValueType{DataTypeSetOfStrings, set}, nil
}

func (ctx *yastCtx) unmarshalSetOfStringsValueFromContent(v interface{}) (*AttributeValueType, error) {
	set, err := ctx.extractContentByItem(v)
	if err != nil || set == nil {
		return nil, err
	}

	return ctx.unmarshalSetOfStringsImmediateValue(set)
}

func (ctx *yastCtx) unmarshalSetOfStringsValue(v interface{}) (*AttributeValueType, error) {
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

func (ctx *yastCtx) unmarshalSetOfNetworksValueItem(v interface{}, i int, set []net.IPNet) ([]net.IPNet, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	s, err := ctx.validateString(v, "element of set of networks")
	if err != nil {
		return nil, err
	}

	_, n, err := net.ParseCIDR(s)
	if err != nil {
		return nil, ctx.errorf("Expected value of network type but got %#v (%v)", s, err)
	}

	return append(set, *n), nil
}

func (ctx *yastCtx) unmarshalSetOfNetworksImmediateValue(v interface{}) (*AttributeValueType, error) {
	items, err := ctx.validateList(v, "")
	if err != nil {
		return nil, nil
	}

	set := make([]net.IPNet, 0)
	for i, item := range items {
		set, err = ctx.unmarshalSetOfNetworksValueItem(item, i, set)
		if err != nil {
			return nil, err
		}
	}

	return &AttributeValueType{DataTypeSetOfNetworks, set}, nil
}

func (ctx *yastCtx) unmarshalSetOfNetworksValueFromContent(v interface{}) (*AttributeValueType, error) {
	set, err := ctx.extractContentByItem(v)
	if err != nil || set == nil {
		return nil, err
	}

	return ctx.unmarshalSetOfNetworksImmediateValue(set)
}

func (ctx *yastCtx) unmarshalSetOfNetworksValue(v interface{}) (*AttributeValueType, error) {
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

func (ctx *yastCtx) unmarshalSetOfDomainsValueItem(v interface{}, i int, set *SetOfSubdomains) error {
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

	set.addToSetOfDomains(d, nil)

	return nil
}

func (ctx *yastCtx) unmarshalSetOfDomainsImmediateValue(v interface{}) (*AttributeValueType, error) {
	items, err := ctx.validateList(v, "")
	if err != nil {
		return nil, nil
	}

	set := SetOfSubdomains{false, nil, make(map[string]*SetOfSubdomains)}
	for i, item := range items {
		err = ctx.unmarshalSetOfDomainsValueItem(item, i, &set)
		if err != nil {
			return nil, err
		}
	}

	return &AttributeValueType{DataTypeSetOfDomains, set}, nil
}

func (ctx *yastCtx) unmarshalSetOfDomainsValueFromContent(v interface{}) (*AttributeValueType, error) {
	set, err := ctx.extractContentByItem(v)
	if err != nil || set == nil {
		return nil, err
	}

	return ctx.unmarshalSetOfDomainsImmediateValue(set)
}

func (ctx *yastCtx) unmarshalSetOfDomainsValue(v interface{}) (*AttributeValueType, error) {
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

func (ctx *yastCtx) unmarshalValueByType(t int, v interface{}) (*AttributeValueType, error) {
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
	}

	return nil, ctx.errorf("Parsing for type %s hasn't been implemented yet", DataTypeNames[t])
}

func (ctx *yastCtx) unmarshalValue(v interface{}) (AttributeValueType, error) {
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

func (ctx *yastCtx) unmarshalArgument(v interface{}, i int, exprs []ExpressionType) ([]ExpressionType, error) {
	ctx.pushNodeSpec("%d", i+1)
	defer ctx.popNodeSpec()

	e, err := ctx.unmarshalExpression(v)
	if err != nil {
		return nil, err
	}

	return append(exprs, e), nil
}

func (ctx *yastCtx) unmarshalArguments(v interface{}) ([]ExpressionType, error) {
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

func (ctx *yastCtx) unmarshalExpression(expr interface{}) (ExpressionType, error) {
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
