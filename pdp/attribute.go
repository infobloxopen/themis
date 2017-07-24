package pdp

import (
	"fmt"
	"net"
	"strings"

	"github.com/infobloxopen/go-trees/domaintree"
	"github.com/infobloxopen/go-trees/iptree"
	"github.com/infobloxopen/go-trees/strtree"
)

const (
	typeUndefined = iota
	typeBoolean
	typeString
	typeAddress
	typeNetwork
	typeDomain
	typeSetOfStrings
	typeSetOfNetworks
	typeSetOfDomains
	typeListOfStrings
)

var typeNames = []string{
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

var undefinedValue = attributeValue{}

type attribute struct {
	id string
	t  int
}

func (a attribute) describe() string {
	return fmt.Sprintf("attr(%s.%s)", a.id, typeNames[a.t])
}

func (a attribute) newMissingError() error {
	return newMissingAttributeError(a.describe())
}

type attributeValue struct {
	t int
	v interface{}
}

func makeBooleanValue(v bool) attributeValue {
	return attributeValue{
		t: typeBoolean,
		v: v}
}

func makeStringValue(v string) attributeValue {
	return attributeValue{
		t: typeString,
		v: v}
}

func makeAddressValue(v net.IP) attributeValue {
	return attributeValue{
		t: typeAddress,
		v: v}
}

func makeNetworkValue(v *net.IPNet) attributeValue {
	return attributeValue{
		t: typeAddress,
		v: v}
}

func makeDomainValue(v string) attributeValue {
	return attributeValue{
		t: typeDomain,
		v: v}
}

func makeSetOfStringsValue(v *strtree.Tree) attributeValue {
	return attributeValue{
		t: typeSetOfStrings,
		v: v}
}

func makeSetOfNetworksValue(v *iptree.Tree) attributeValue {
	return attributeValue{
		t: typeSetOfNetworks,
		v: v}
}

func makeSetOfDomainsValue(v *domaintree.Node) attributeValue {
	return attributeValue{
		t: typeSetOfDomains,
		v: v}
}

func makeListOfStringsValue(v []string) attributeValue {
	return attributeValue{
		t: typeListOfStrings,
		v: v}
}

func (v attributeValue) describe() string {
	switch v.t {
	case typeUndefined:
		return "val(undefined)"

	case typeBoolean:
		return fmt.Sprintf("%v", v.v.(bool))

	case typeString:
		return fmt.Sprintf("%q", v.v.(string))

	case typeAddress:
		return v.v.(net.IP).String()

	case typeNetwork:
		return v.v.(*net.IPNet).String()

	case typeDomain:
		return fmt.Sprintf("domain(%s)", v.v.(string))

	case typeSetOfStrings:
		s := []string{}
		for p := range v.v.(*strtree.Tree).Enumerate() {
			s = append(s, fmt.Sprintf("%q", p.Key))
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("set(%s)", strings.Join(s, ", "))

	case typeSetOfNetworks:
		s := []string{}
		for p := range v.v.(*iptree.Tree).Enumerate() {
			s = append(s, p.Key.String())
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("set(%s)", strings.Join(s, ", "))

	case typeSetOfDomains:
		s := []string{}
		for p := range v.v.(*domaintree.Node).Enumerate() {
			s = append(s, fmt.Sprintf("%q", p.Key))
			if len(s) > 2 {
				s[2] = "..."
				break
			}
		}

		return fmt.Sprintf("domains(%s)", strings.Join(s, ", "))

	case typeListOfStrings:
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

func (v attributeValue) typeHRName() string {
	return typeNames[v.t]
}

func (v attributeValue) typeCheck(t int) error {
	if v.t != t {
		return newAttributeValueTypeError(t, v.t, v.describe())
	}

	return nil
}

func (v attributeValue) boolean() (bool, error) {
	err := v.typeCheck(typeBoolean)
	if err != nil {
		return false, err
	}

	return v.v.(bool), nil
}

func (v attributeValue) str() (string, error) {
	err := v.typeCheck(typeString)
	if err != nil {
		return "", err
	}

	return v.v.(string), nil
}

func (v attributeValue) address() (net.IP, error) {
	err := v.typeCheck(typeAddress)
	if err != nil {
		return nil, err
	}

	return v.v.(net.IP), nil
}

func (v attributeValue) network() (*net.IPNet, error) {
	err := v.typeCheck(typeNetwork)
	if err != nil {
		return nil, err
	}

	return v.v.(*net.IPNet), nil
}

func (v attributeValue) domain() (string, error) {
	err := v.typeCheck(typeDomain)
	if err != nil {
		return "", err
	}

	return v.v.(string), nil
}

func (v attributeValue) setOfStrings() (*strtree.Tree, error) {
	err := v.typeCheck(typeSetOfStrings)
	if err != nil {
		return nil, err
	}

	return v.v.(*strtree.Tree), nil
}

func (v attributeValue) setOfNetworks() (*iptree.Tree, error) {
	err := v.typeCheck(typeSetOfNetworks)
	if err != nil {
		return nil, err
	}

	return v.v.(*iptree.Tree), nil
}

func (v attributeValue) setOfDomains() (*domaintree.Node, error) {
	err := v.typeCheck(typeSetOfDomains)
	if err != nil {
		return nil, err
	}

	return v.v.(*domaintree.Node), nil
}

func (v attributeValue) listOfStrings() ([]string, error) {
	err := v.typeCheck(typeListOfStrings)
	if err != nil {
		return nil, err
	}

	return v.v.([]string), nil
}

func (v attributeValue) calculate(ctx *Context) (attributeValue, error) {
	return v, nil
}

type attributeAssignmentExpression struct {
	a attribute
	e expression
}

type attributeDesignator struct {
	a attribute
}

func (d attributeDesignator) calculate(ctx *Context) (attributeValue, error) {
	return ctx.getAttribute(d.a)
}

func newStrTree(args ...string) *strtree.Tree {
	t := strtree.NewTree()
	for i, s := range args {
		t.InplaceInsert(s, i)
	}

	return t
}
