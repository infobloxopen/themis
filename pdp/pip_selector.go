package pdp

import (
	"context"
	"fmt"
	"strings"

	ps "github.com/infobloxopen/themis/pip-service"
)

const pIPServicePort = ":5356"

type PIPSelector struct {
	service string

	queryType string
	path      []Expression
	t         int
}

func MakePIPSelector(service, queryType string, path []Expression, t int) (PIPSelector, error) {
	return PIPSelector{
		service:   service,
		queryType: queryType,
		path:      path,
		t:         t}, nil
}

func (s PIPSelector) GetResultType() int {
	return s.t
}

func (s PIPSelector) describe() string {
	return fmt.Sprintf("PIPselector(%s.%s)", s.service, s.queryType)
}

func (s PIPSelector) getAttributeValue(ctx *Context) (AttributeValue, error) {
	attrList := []*ps.Attribute{}
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
		attr := ps.Attribute{Id: id, Type: int32(t), Value: serializedVal}
		attrList = append(attrList, &attr)
	}

	conn, err := ctx.p.GetConnection(s.service)
	if err != nil {
		return undefinedValue, bindError(fmt.Errorf("cannot connect to pip, err: '%s'", err), s.describe())
	}
	c := ps.NewPIPClient(conn)
	request := &ps.Request{QueryType: s.queryType, Attributes: attrList}

	r, err := c.GetAttribute(context.Background(), request)
	if err != nil {
		return undefinedValue, bindError(err, s.describe())
	}
	if r.Status != ps.Response_OK {
		return undefinedValue, bindError(fmt.Errorf("Unexpected response status '%s'", r.Status), s.describe())
	}
	res := r.GetValues()
	// fmt.Printf("res = '%v'\n", res)
	// For now, assume response only contain one attribute
	val := res[0].GetValue()
	t := int(res[0].GetType())
	// fmt.Printf("PIP returned value='%v, type = %d'\n", val, t)
	if s.t != t {
		return undefinedValue, bindError(fmt.Errorf("Unexpected response value type '%d', expecting '%d'", t, s.t), s.describe())
	}

	switch s.t {
	case TypeString:
		return MakeStringValue(val), nil

	case TypeListOfStrings:
		return MakeListOfStringsValue(strings.Split(val, ",")), nil
	}

	return undefinedValue, bindError(fmt.Errorf("Unexpected response value type '%d'", t), s.describe())
}

func (s PIPSelector) calculate(ctx *Context) (AttributeValue, error) {
	return s.getAttributeValue(ctx)
}
