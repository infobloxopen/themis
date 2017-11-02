package pdp_service

// Effect field values
const (
	DENY            = 0
	PERMIT          = 1
	NOTAPPLICABLE   = 2
	INDETERMINATE   = 3
	INDETERMINATED  = 4
	INDETERMINATEP  = 5
	INDETERMINATEDP = 6
)

func EffectName(effect byte) string {
	switch effect {
	case 0:
		return "DENY"
	case 1:
		return "PERMIT"
	case 2:
		return "NOTAPPLICABLE"
	case 3:
		return "INDETERMINATE"
	case 4:
		return "INDETERMINATED"
	case 5:
		return "INDETERMINATEP"
	case 6:
		return "INDETERMINATEDP"
	}
	return "INVALID EFFECT"
}

type Attribute struct {
	Id    string
	Type  string
	Value string
}

type Request struct {
	Attributes []*Attribute
}

type Response struct {
	Effect      byte
	Reason      string
	Obligations []*Attribute
}

func MarshalRequest(r *Request) []byte {
	count := len(r.Attributes)
	length := 1
	for j := 0; j < count; j++ {
		length += 3 + len(r.Attributes[j].Id) + len(r.Attributes[j].Type) + len(r.Attributes[j].Value)
	}
	ret := make([]byte, length)
	i := 0
	ret[i] = byte(count)
	i++
	for j := 0; j < count; j++ {
		lId := len(r.Attributes[j].Id)
		ret[i] = byte(lId)
		i++
		for k := 0; k < lId; k++ {
			ret[i] = r.Attributes[j].Id[k]
			i++
		}
		lType := len(r.Attributes[j].Type)
		ret[i] = byte(lType)
		i++
		for k := 0; k < lType; k++ {
			ret[i] = r.Attributes[j].Type[k]
			i++
		}
		lValue := len(r.Attributes[j].Value)
		ret[i] = byte(lValue)
		i++
		for k := 0; k < lValue; k++ {
			ret[i] = r.Attributes[j].Value[k]
			i++
		}
	}
	return ret
}

func MarshalResponse(r *Response) []byte {
	count := len(r.Obligations)
	rLen := len(r.Reason)
	length := 3 + rLen
	for j := 0; j < count; j++ {
		length += 3 + len(r.Obligations[j].Id) + len(r.Obligations[j].Type) + len(r.Obligations[j].Value)
	}
	ret := make([]byte, length)
	i := 0
	ret[i] = r.Effect
	i++
	ret[i] = byte(count)
	i++
	ret[i] = byte(rLen)
	i++
	for j := 0; j < rLen; j++ {
		ret[i] = r.Reason[j]
		i++
	}
	for j := 0; j < count; j++ {
		lId := len(r.Obligations[j].Id)
		ret[i] = byte(lId)
		i++
		for k := 0; k < lId; k++ {
			ret[i] = r.Obligations[j].Id[k]
			i++
		}
		lType := len(r.Obligations[j].Type)
		ret[i] = byte(lType)
		i++
		for k := 0; k < lType; k++ {
			ret[i] = r.Obligations[j].Type[k]
			i++
		}
		lValue := len(r.Obligations[j].Value)
		ret[i] = byte(lValue)
		i++
		for k := 0; k < lValue; k++ {
			ret[i] = r.Obligations[j].Value[k]
			i++
		}
	}
	return ret
}

func UnmarshalRequest(data []byte) *Request {
	i := 0
	count := data[i]
	i++
	ret := &Request{
		Attributes: make([]*Attribute, count),
	}
	for j := 0; j < int(count); j++ {
		sId := i + 1
		eId := int(data[i]) + sId
		sType := eId + 1
		eType := int(data[eId]) + sType
		sValue := eType + 1
		eValue := int(data[eType]) + sValue
		ret.Attributes[j] = &Attribute{
			Id:    string(data[sId:eId]),
			Type:  string(data[sType:eType]),
			Value: string(data[sValue:eValue]),
		}
		i = eValue
	}
	return ret
}

func UnmarshalResponse(data []byte) *Response {
	i := 0
	effect := data[i]
	i++
	count := data[i]
	i++
	rLen := int(data[i])
	i++
	ret := &Response{
		Effect:      effect,
		Reason:      string(data[i : i+rLen]),
		Obligations: make([]*Attribute, count),
	}
	i += rLen
	for j := 0; j < int(count); j++ {
		sId := i + 1
		eId := int(data[i]) + sId
		sType := eId + 1
		eType := int(data[eId]) + sType
		sValue := eType + 1
		eValue := int(data[eType]) + sValue
		ret.Obligations[j] = &Attribute{
			Id:    string(data[sId:eId]),
			Type:  string(data[sType:eType]),
			Value: string(data[sValue:eValue]),
		}
		i = eValue
	}
	return ret
}
