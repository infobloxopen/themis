package pkg

import (
	"fmt"
	"strings"
)

func makeArgParsers(args []string, z string) []string {
	if len(args) <= 0 {
		return nil
	}

	out := make([]string, 0, len(args))

	last := len(args) - 1
	for i, a := range args {
		out = append(out, makeArgParser(i, a, z, i >= last))
	}

	return out
}

func makeArgParser(i int, t, z string, last bool) string {
	buf := "in"
	if last {
		buf = "_"
	}

	return fmt.Sprintf("\tv%d, %s, err := %s(in)\n"+
		"\tif err != nil {\n"+
		"\t\treturn %s, err\n"+
		"\t}\n", i, buf, parserNameMap[strings.ToLower(t)], z)
}

const (
	booleanParserName       = "pdp.GetInfoRequestBooleanValue"
	stringParserName        = "pdp.GetInfoRequestStringValue"
	integerParserName       = "pdp.GetInfoRequestIntegerValue"
	floatParserName         = "pdp.GetInfoRequestFloatValue"
	addressParserName       = "pdp.GetInfoRequestAddressValue"
	networkParserName       = "pdp.GetInfoRequestNetworkValue"
	domainParserName        = "pdp.GetInfoRequestDomainValue"
	setOfStringsParserName  = "pdp.GetInfoRequestSetOfStringsValue"
	setOfNetworksParserName = "pdp.GetInfoRequestSetOfNetworksValue"
	setOfDomainsParserName  = "pdp.GetInfoRequestSetOfDomainsValue"
	listOfStringsParserName = "pdp.GetInfoRequestListOfStringsValue"
)

var parserNameMap = map[string]string{
	pipTypeBoolean:       booleanParserName,
	pipTypeString:        stringParserName,
	pipTypeInteger:       integerParserName,
	pipTypeFloat:         floatParserName,
	pipTypeAddress:       addressParserName,
	pipTypeNetwork:       networkParserName,
	pipTypeDomain:        domainParserName,
	pipTypeSetOfStrings:  setOfStringsParserName,
	pipTypeSetOfNetworks: setOfNetworksParserName,
	pipTypeSetOfDomains:  setOfDomainsParserName,
	pipTypeListOfStrings: listOfStringsParserName,
}
