package pdp

import (
	"fmt"
	"net"
	"strings"
)

const (
	DataTypeUndefined = iota
	DataTypeBoolean
	DataTypeString
	DataTypeAddress
	DataTypeNetwork
	DataTypeDomain
	DataTypeSetOfStrings
	DataTypeSetOfNetworks
	DataTypeSetOfDomains
)

var DataTypeNames map[int]string = map[int]string{
	DataTypeUndefined:     yastTagDataTypeUndefined,
	DataTypeBoolean:       yastTagDataTypeBoolean,
	DataTypeString:        yastTagDataTypeString,
	DataTypeAddress:       yastTagDataTypeAddress,
	DataTypeNetwork:       yastTagDataTypeNetwork,
	DataTypeDomain:        yastTagDataTypeDomain,
	DataTypeSetOfStrings:  yastTagDataTypeSetOfStrings,
	DataTypeSetOfNetworks: yastTagDataTypeSetOfNetworks,
	DataTypeSetOfDomains:  yastTagDataTypeSetOfDomains}

var DataTypeIDs map[string]int = map[string]int{
	yastTagDataTypeUndefined:     DataTypeUndefined,
	yastTagDataTypeBoolean:       DataTypeBoolean,
	yastTagDataTypeString:        DataTypeString,
	yastTagDataTypeAddress:       DataTypeAddress,
	yastTagDataTypeNetwork:       DataTypeNetwork,
	yastTagDataTypeDomain:        DataTypeDomain,
	yastTagDataTypeSetOfStrings:  DataTypeSetOfStrings,
	yastTagDataTypeSetOfNetworks: DataTypeSetOfNetworks,
	yastTagDataTypeSetOfDomains:  DataTypeSetOfDomains}

type AttributeType struct {
	ID       string
	DataType int
}

type AttributeValueType struct {
	DataType int
	Value    interface{}
}

func (v AttributeValueType) describe() string {
	switch v.DataType {
	case DataTypeUndefined:
		return yastTagDataTypeUndefined

	case DataTypeBoolean:
		return fmt.Sprintf("%v", v.Value)

	case DataTypeString:
		return fmt.Sprintf("%q", v.Value)

	case DataTypeAddress:
		return v.Value.(net.IP).String()

	case DataTypeNetwork:
		n := v.Value.(net.IPNet)
		return n.String()

	case DataTypeDomain:
		return fmt.Sprintf("%q", v.Value)

	case DataTypeSetOfStrings:
		items := []string{}
		i := 0
		for s := range v.Value.(map[string]bool) {
			if i > 1 {
				items = append(items, "...")
				break
			}

			items = append(items, s)
			i++
		}

		return fmt.Sprintf("[%s]", strings.Join(items, ", "))

	case DataTypeSetOfNetworks:
		items := []string{}
		i := 0
		for n := range v.Value.(*SetOfNetworks).Iterate() {
			if i > 1 {
				items = append(items, "...")
				break
			}

			items = append(items, n.Network.String())
			i++
		}

		return fmt.Sprintf("[%s]", strings.Join(items, ", "))

	case DataTypeSetOfDomains:
		items := []string{}
		i := 0
		for d := range v.Value.(*SetOfSubdomains).Iterate() {
			if i > 1 {
				items = append(items, "...")
				break
			}

			items = append(items, d.Domain)
			i++
		}

		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	}

	return fmt.Sprintf("<unknown value type %d>", DataTypeSetOfDomains)
}

func (v AttributeValueType) getResultType() int {
	return v.DataType
}

func (v AttributeValueType) calculate(ctx *Context) (AttributeValueType, error) {
	return v, nil
}

type AttributeAssignmentExpressionType struct {
	Attribute  AttributeType
	Expression ExpressionType
}

type AttributeDesignatorType struct {
	Attribute AttributeType
}

func (d AttributeDesignatorType) describe() string {
	return fmt.Sprintf("%s(%s)", yastTagAttribute, d.Attribute.ID)
}

func (d AttributeDesignatorType) getResultType() int {
	return d.Attribute.DataType
}

func (d AttributeDesignatorType) calculate(ctx *Context) (AttributeValueType, error) {
	return ctx.GetAttribute(d.Attribute)
}

func ExtractBooleanValue(v AttributeValueType, desc string) (bool, error) {
	if v.DataType != DataTypeBoolean {
		return false, fmt.Errorf("Expected %s as %s but got %s",
			DataTypeNames[DataTypeBoolean], desc, DataTypeNames[v.DataType])
	}

	return v.Value.(bool), nil
}

func ExtractStringValue(v AttributeValueType, desc string) (string, error) {
	if v.DataType != DataTypeString {
		return "", fmt.Errorf("Expected %s as %s but got %s",
			DataTypeNames[DataTypeString], desc, DataTypeNames[v.DataType])
	}

	return v.Value.(string), nil
}

func ExtractAddressValue(v AttributeValueType, desc string) (net.IP, error) {
	if v.DataType != DataTypeAddress {
		return net.IP{}, fmt.Errorf("Expected %s as %s but got %s",
			DataTypeNames[DataTypeAddress], desc, DataTypeNames[v.DataType])
	}

	return v.Value.(net.IP), nil
}

func ExtractNetworkValue(v AttributeValueType, desc string) (net.IPNet, error) {
	if v.DataType != DataTypeNetwork {
		return net.IPNet{}, fmt.Errorf("Expected %s as %s but got %s",
			DataTypeNames[DataTypeNetwork], desc, DataTypeNames[v.DataType])
	}

	return v.Value.(net.IPNet), nil
}

func ExtractDomainValue(v AttributeValueType, desc string) (string, error) {
	if v.DataType != DataTypeDomain {
		return "", fmt.Errorf("Expected %s as %s but got %s",
			DataTypeNames[DataTypeDomain], desc, DataTypeNames[v.DataType])
	}

	return v.Value.(string), nil
}

func ExtractSetOfStringsValue(v AttributeValueType, desc string) (map[string]bool, error) {
	if v.DataType != DataTypeSetOfStrings {
		return nil, fmt.Errorf("Expected %s as %s but got %s",
			DataTypeNames[DataTypeSetOfStrings], desc, DataTypeNames[v.DataType])
	}

	return v.Value.(map[string]bool), nil
}

func ExtractSetOfNetworksValue(v AttributeValueType, desc string) (*SetOfNetworks, error) {
	if v.DataType != DataTypeSetOfNetworks {
		return nil, fmt.Errorf("Expected %s as %s but got %s",
			DataTypeNames[DataTypeSetOfNetworks], desc, DataTypeNames[v.DataType])
	}

	return v.Value.(*SetOfNetworks), nil
}

func ExtractSetOfDomainsValue(v AttributeValueType, desc string) (*SetOfSubdomains, error) {
	if v.DataType != DataTypeSetOfDomains {
		return nil, fmt.Errorf("Expected %s as %s but got %s",
			DataTypeNames[DataTypeSetOfDomains], desc, DataTypeNames[v.DataType])
	}

	return v.Value.(*SetOfSubdomains), nil
}
