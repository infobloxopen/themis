package pkg

import (
	"fmt"
	"strings"
)

func makeArgParsers(args []string, z string) ([]string, error) {
	out := make([]string, 0, len(args))
	for i, a := range args {
		p, err := makeArgParser(i, a, z)
		if err != nil {
			return nil, err
		}

		out = append(out, p)
	}

	return out, nil
}

func makeArgParser(i int, t, z string) (string, error) {
	n, ok := parserNameMap[strings.ToLower(t)]
	if !ok {
		return "", fmt.Errorf("argument %d: unknown type %q", i, t)
	}

	return fmt.Sprintf("\tv%d, in, err := %s(in)\n"+
		"\tif err != nil {\n"+
		"\t\treturn %s, err\n"+
		"\t}\n", i, n, z), nil
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
