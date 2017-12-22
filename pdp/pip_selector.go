package pdp

import (
	"fmt"
	"strings"

	info "github.com/infobloxopen/themis/pdp-info"
)

type PipSelector struct {
	uri  string
	path []Expression
	t    int
}

func MakePipSelector(uri string, path []Expression, t int) PipSelector {
	return PipSelector{
		uri:  uri,
		path: path,
		t:    t}
}

func (s PipSelector) GetResultType() int {
	return s.t
}

func (s PipSelector) describe() string {
	return fmt.Sprintf("PipSelector(%s)", s.uri)
}

func (s PipSelector) calculate(ctx *Context) (AttributeValue, error) {
	attrList := []*info.Attribute{}
	for _, p := range s.path {
		val, err := p.calculate(ctx)
		if err != nil {
			return undefinedValue, bindError(err, s.describe())
		}
		id := p.(AttributeDesignator).a.id
		t := p.(AttributeDesignator).a.t
		serializedVal, err := val.Serialize()
		if err != nil {
			return undefinedValue, bindError(err, s.describe())
		}
		attr := info.Attribute{Id: id, Type: TypeKeys[t], Value: serializedVal}
		attrList = append(attrList, &attr)
	}

	pip := info.GetPipService()

	res, err := pip.Get(s.uri, attrList)
	if err != nil {
		return undefinedValue, bindError(err, s.describe())
	}

	if len(res) != 1 {
		err := fmt.Errorf("Expected exactly one attribute, got: %d", len(res))
		return undefinedValue, bindError(err, s.describe())
	}

	val := res[0].GetValue()
	t := TypeIDs[res[0].GetType()]

	if t != s.t {
		return undefinedValue, bindError(newInvalidPipResponseTypeError(s.t, t), s.describe())
	}

	return MakeListOfStringsValue(strings.Split(val, ",")), nil
}
