package pdp

import (
	"context"
	"fmt"

	ps "github.com/infobloxopen/themis/pip-service"
	"google.golang.org/grpc"
)

const pIPServicePort = ":5356"

type PIPSelector struct {
	serviceAddr string
	queryType string
	path      []Expression
	t         int
}


func MakePIPSelector(service, queryType string, path []Expression, t int) (PIPSelector, error) {
	addr, err := ps.GetPIPConnection(service)
	if err != nil {
		return PIPSelector{}, err
	}

	return PIPSelector{
		serviceAddr: addr,
		queryType: queryType,
		path:      path,
		t:         t}, nil
}

func (s PIPSelector) GetResultType() int {
	return s.t
}

func (s PIPSelector) describe() string {
	return fmt.Sprintf("PIPselector(%s.%s)", s.serviceAddr, s.queryType)
}

func (s PIPSelector) getAttributeValue(ctx *Context) (AttributeValue, error) {
	domainAttr, err := s.path[0].calculate(ctx)

	domainStr, err := domainAttr.Serialize()
	if err != nil {
		return undefinedValue, bindError(err, s.describe())
	}
	fmt.Printf("domainStr is '%v'\n", domainStr)

	conn, err := grpc.Dial(s.serviceAddr, grpc.WithInsecure())
	if err != nil {
		return undefinedValue, bindError(fmt.Errorf("cannot connect to pip: %s\n", err), s.describe())
	}
	defer conn.Close()

	c := ps.NewPIPClient(conn)
	reqAttr := &ps.Attribute{Value: domainStr}
	request := &ps.Request{QueryType: s.queryType, Attributes: []*ps.Attribute{reqAttr}}

	r, err := c.GetAttribute(context.Background(), request)
	if err != nil {
		return undefinedValue, bindError(err, s.describe())
	}
	if r.Status != ps.Response_OK {
		return undefinedValue, bindError(fmt.Errorf("Unexpected response status '%s'", r.Status), s.describe())
	}
	res := r.GetValues()
	val := res[0].GetValue()

	fmt.Printf("PIP returned value='%v'\n", val)

	return MakeStringValue(val), nil
}

func (s PIPSelector) calculate(ctx *Context) (AttributeValue, error) {
	return s.getAttributeValue(ctx)
}
