package pdp

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/idna"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

const (
	TypeUndefined = iota
	TypeBoolean
	TypeString
	TypeAddress
	TypeNetwork
	TypeDomain
	TypeSetOfStrings
	TypeSetOfNetworks
	TypeSetOfDomains
	TypeListOfStrings

	typesTotal
)

var (
	TypeNames = []string{
		"Undefined",
		"Boolean",
		"String",
		"Address",
		"Network",
		"Domain",
		"Set of Strings",
		"Set of Networks",
		"Set of Domains",
		"List of Strings"}

	TypeKeys = []string{}
	TypeIDs  = map[string]int{}
)

var undefinedValue = AttributeValue{}

func init() {
	for t := 0; t < typesTotal; t++ {
		key := strings.ToLower(TypeNames[t])
		TypeKeys = append(TypeKeys, key)
		TypeIDs[key] = t
	}
}

type Attribute struct {
	id string
	t  int
}

func MakeAttribute(ID string, t int) Attribute {
	return Attribute{id: ID, t: t}
}

func (a Attribute) GetType() int {
	return a.t
}

func (a Attribute) describe() string {
	return fmt.Sprintf("attr(%s.%s)", a.id, TypeNames[a.t])
}

type AttributeValue struct {
	t int
	v interface{}
}

func MakeBooleanValue(v bool) AttributeValue {
	return AttributeValue{
		t: TypeBoolean,
		v: v}
}

func MakeStringValue(v string) AttributeValue {
	return AttributeValue{
		t: TypeString,
		v: v}
}

func MakeAddressValue(v net.IP) AttributeValue {
	return AttributeValue{
		t: TypeAddress,
		v: v}
}

func MakeNetworkValue(v *net.IPNet) AttributeValue {
	return AttributeValue{
		t: TypeNetwork,
		v: v}
}

func MakeDomainValue(v string) AttributeValue {
	return AttributeValue{
		t: TypeDomain,
		v: v}
}

func MakeSetOfStringsValue(v *strtree.Tree) AttributeValue {
	return AttributeValue{
		t: TypeSetOfStrings,
		v: v}
}

func MakeSetOfNetworksValue(v *iptree.Tree) AttributeValue {
	return AttributeValue{
		t: TypeSetOfNetworks,
		v: v}
}

func MakeSetOfDomainsValue(v *domaintree.Node) AttributeValue {
	return AttributeValue{
		t: TypeSetOfDomains,
		v: v}
}

func MakeListOfStringsValue(v []string) AttributeValue {
	return AttributeValue{
		t: TypeListOfStrings,
		v: v}
}

func MakeValueFromString(t int, s string) (AttributeValue, error) {
	switch t {
	case TypeUndefined:
		return undefinedValue, newInvalidTypeStringCastError(t)

	case TypeSetOfStrings, TypeSetOfNetworks, TypeSetOfDomains, TypeListOfStrings:
		return undefinedValue, newNotImplementedStringCastError(t)

	case TypeBoolean:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return undefinedValue, newInvalidBooleanStringCastError(s, err)
		}

		return MakeBooleanValue(b), nil

	case TypeString:
		return MakeStringValue(s), nil

	case TypeAddress:
		a := net.ParseIP(s)
		if a == nil {
			return undefinedValue, newInvalidAddressStringCastError(s)
		}

		return MakeAddressValue(a), nil

	case TypeNetwork:
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return undefinedValue, newInvalidNetworkStringCastError(s, err)
		}

		return MakeNetworkValue(n), nil

	case TypeDomain:
		return MakeDomainValue(s), nil
	}

	return undefinedValue, newUnknownTypeStringCastError(t)
}

func (v AttributeValue) GetResultType() int {
	return v.t
}

func (v AttributeValue) describe() string {
	switch v.t {
	case TypeUndefined:
		return "val(undefined)"

	case TypeBoolean:
		return fmt.Sprintf("%v", v.v.(bool))

	case TypeString:
		return fmt.Sprintf("%q", v.v.(string))

	case TypeAddress:
		return v.v.(net.IP).String()

	case TypeNetwork:
		return v.v.(*net.IPNet).String()

	case TypeDomain:
		return fmt.Sprintf("domain(%s)", v.v.(string))

	case TypeSetOfStrings:
		s := []string{}
		for p := range v.v.(*strtree.Tree).Enumerate() {
			s = append(s, fmt.Sprintf("%q", p.Key))
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("set(%s)", strings.Join(s, ", "))

	case TypeSetOfNetworks:
		s := []string{}
		for p := range v.v.(*iptree.Tree).Enumerate() {
			s = append(s, p.Key.String())
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("set(%s)", strings.Join(s, ", "))

	case TypeSetOfDomains:
		s := []string{}
		for p := range v.v.(*domaintree.Node).Enumerate() {
			s = append(s, fmt.Sprintf("%q", p.Key))
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("domains(%s)", strings.Join(s, ", "))

	case TypeListOfStrings:
		s := []string{}
		for _, item := range v.v.([]string) {
			s = append(s, fmt.Sprintf("%q", item))
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("[%s]", strings.Join(s, ", "))
	}

	return "val(unknown type)"
}

func (v AttributeValue) typeCheck(t int) error {
	if v.t != t {
		return bindError(newAttributeValueTypeError(t, v.t), v.describe())
	}

	return nil
}

func (v AttributeValue) boolean() (bool, error) {
	err := v.typeCheck(TypeBoolean)
	if err != nil {
		return false, err
	}

	return v.v.(bool), nil
}

func (v AttributeValue) str() (string, error) {
	err := v.typeCheck(TypeString)
	if err != nil {
		return "", err
	}

	return v.v.(string), nil
}

func (v AttributeValue) address() (net.IP, error) {
	err := v.typeCheck(TypeAddress)
	if err != nil {
		return nil, err
	}

	return v.v.(net.IP), nil
}

func (v AttributeValue) network() (*net.IPNet, error) {
	err := v.typeCheck(TypeNetwork)
	if err != nil {
		return nil, err
	}

	return v.v.(*net.IPNet), nil
}

func (v AttributeValue) domain() (string, error) {
	err := v.typeCheck(TypeDomain)
	if err != nil {
		return "", err
	}

	return v.v.(string), nil
}

func (v AttributeValue) setOfStrings() (*strtree.Tree, error) {
	err := v.typeCheck(TypeSetOfStrings)
	if err != nil {
		return nil, err
	}

	return v.v.(*strtree.Tree), nil
}

func (v AttributeValue) setOfNetworks() (*iptree.Tree, error) {
	err := v.typeCheck(TypeSetOfNetworks)
	if err != nil {
		return nil, err
	}

	return v.v.(*iptree.Tree), nil
}

func (v AttributeValue) setOfDomains() (*domaintree.Node, error) {
	err := v.typeCheck(TypeSetOfDomains)
	if err != nil {
		return nil, err
	}

	return v.v.(*domaintree.Node), nil
}

func (v AttributeValue) listOfStrings() ([]string, error) {
	err := v.typeCheck(TypeListOfStrings)
	if err != nil {
		return nil, err
	}

	return v.v.([]string), nil
}

func (v AttributeValue) calculate(ctx *Context) (AttributeValue, error) {
	return v, nil
}

func (v AttributeValue) Serialize() (string, error) {
	switch v.t {
	case TypeUndefined:
		return "", newInvalidTypeSerializationError(v.t)

	case TypeBoolean:
		return strconv.FormatBool(v.v.(bool)), nil

	case TypeString:
		return v.v.(string), nil

	case TypeAddress:
		return v.v.(net.IP).String(), nil

	case TypeNetwork:
		return v.v.(*net.IPNet).String(), nil

	case TypeDomain:
		return v.v.(string), nil

	case TypeSetOfStrings:
		s := sortSetOfStrings(v.v.(*strtree.Tree))
		for i, item := range s {
			s[i] = strconv.Quote(item)
		}

		return strings.Join(s, ","), nil

	case TypeSetOfNetworks:
		s := []string{}
		for p := range v.v.(*iptree.Tree).Enumerate() {
			s = append(s, strconv.Quote(p.Key.String()))
		}

		return strings.Join(s, ","), nil

	case TypeSetOfDomains:
		s := []string{}
		for p := range v.v.(*domaintree.Node).Enumerate() {
			s = append(s, strconv.Quote(p.Key))
		}

		return strings.Join(s, ","), nil

	case TypeListOfStrings:
		s := []string{}
		for _, item := range v.v.([]string) {
			s = append(s, strconv.Quote(item))
		}

		return strings.Join(s, ","), nil
	}

	return "", newUnknownTypeSerializationError(v.t)
}

type AttributeAssignmentExpression struct {
	a Attribute
	e Expression
}

func MakeAttributeAssignmentExpression(a Attribute, e Expression) AttributeAssignmentExpression {
	return AttributeAssignmentExpression{
		a: a,
		e: e}
}

func (a AttributeAssignmentExpression) Serialize(ctx *Context) (string, string, string, error) {
	ID := a.a.id
	typeName := TypeKeys[a.a.t]

	v, err := a.e.calculate(ctx)
	if err != nil {
		return ID, typeName, "", bindErrorf(err, "assignment to %q", ID)
	}

	t := v.GetResultType()
	if a.a.t != t {
		return ID, typeName, "", bindErrorf(newAssignmentTypeMismatch(a.a, t), "assignment to %q", ID)
	}

	s, err := v.Serialize()
	if err != nil {
		return ID, typeName, "", bindErrorf(err, "assignment to %q", ID)
	}

	return ID, typeName, s, nil
}

type AttributeDesignator struct {
	a Attribute
}

func MakeAttributeDesignator(a Attribute) AttributeDesignator {
	return AttributeDesignator{a}
}

func (d AttributeDesignator) GetResultType() int {
	return d.a.t
}

func (d AttributeDesignator) calculate(ctx *Context) (AttributeValue, error) {
	return ctx.getAttribute(d.a)
}

var domainRegexp = regexp.MustCompile("^[-._a-z0-9]+$")

func AdjustDomainName(s string) (string, error) {
	tmp, err := idna.Punycode.ToASCII(s)
	if err != nil {
		return "", fmt.Errorf("Cannot convert domain [%s]", s)
	}
	ret := strings.ToLower(tmp)
	if !domainRegexp.MatchString(ret) {
		return "", fmt.Errorf("Cannot validate domain [%s]", s)
	}
	return ret, nil
}
