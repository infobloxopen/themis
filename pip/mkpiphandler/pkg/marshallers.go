package pkg

const (
	pdpMarshallerBoolean       = "pdp.MarshalInfoResponseBoolean"
	pdpMarshallerString        = "pdp.MarshalInfoResponseString"
	pdpMarshallerInteger       = "pdp.MarshalInfoResponseInteger"
	pdpMarshallerFloat         = "pdp.MarshalInfoResponseFloat"
	pdpMarshallerAddress       = "pdp.MarshalInfoResponseAddress"
	pdpMarshallerNetwork       = "pdp.MarshalInfoResponseNetwork"
	pdpMarshallerDomain        = "pdp.MarshalInfoResponseDomain"
	pdpMarshallerSetOfStrings  = "pdp.MarshalInfoResponseSetOfStrings"
	pdpMarshallerSetOfNetworks = "pdp.MarshalInfoResponseSetOfNetworks"
	pdpMarshallerSetOfDomains  = "pdp.MarshalInfoResponseSetOfDomains"
	pdpMarshallerListOfStrings = "pdp.MarshalInfoResponseListOfStrings"
)

var marshallerMap = map[string]string{
	pipTypeBoolean:       pdpMarshallerBoolean,
	pipTypeString:        pdpMarshallerString,
	pipTypeInteger:       pdpMarshallerInteger,
	pipTypeFloat:         pdpMarshallerFloat,
	pipTypeAddress:       pdpMarshallerAddress,
	pipTypeNetwork:       pdpMarshallerNetwork,
	pipTypeDomain:        pdpMarshallerDomain,
	pipTypeSetOfStrings:  pdpMarshallerSetOfStrings,
	pipTypeSetOfNetworks: pdpMarshallerSetOfNetworks,
	pipTypeSetOfDomains:  pdpMarshallerSetOfDomains,
	pipTypeListOfStrings: pdpMarshallerListOfStrings,
}
