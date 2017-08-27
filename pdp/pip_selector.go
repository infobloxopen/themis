package pdp

import (
	"context"
	"fmt"

	pip_service "github.com/infobloxopen/themis/pip-service"
	"google.golang.org/grpc"
)

const pIPServicePort = ":5356"

type PIPSelector struct {
	service   string
	queryType string
	path      []Expression
	t         int
}

func MakePIPSelector(service, queryType string, path []Expression, t int) PIPSelector {
	return PIPSelector{
		service:   service + pIPServicePort,
		queryType: queryType,
		path:      path,
		t:         t}
}

func (s PIPSelector) GetResultType() int {
	return s.t
}

func (s PIPSelector) describe() string {
	return fmt.Sprintf("PIPselector(%s.%s)", s.service, s.queryType)
}

func (s PIPSelector) getAttributeValue(ctx *Context) (AttributeValue, error) {
	domainAttr, err := s.path[0].calculate(ctx)

	domainStr, err := domainAttr.Serialize()

	if err != nil {
		return undefinedValue, err
	}
	fmt.Printf("domainStr is '%v'\n", domainStr)

	conn, err := grpc.Dial(s.service, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("cannot connect to pip: %s\n", err)
	}
	defer conn.Close()

	c := pip_service.NewPIPClient(conn)
	reqAttr := &pip_service.Attribute{Value: domainStr}
	request := &pip_service.Request{Attributes: []*pip_service.Attribute{reqAttr}}

	r, err := c.GetAttribute(context.Background(), request)
	if err != nil {
		fmt.Printf("Cannot get value from PIP\n")
		return undefinedValue, err
	}
	res := r.GetValues()
	val := res[0].GetValue()
	fmt.Printf("PIP returned value='%v'\n", val)
	return MakeStringValue(val), nil
	//	return MakeStringValue("Naughty"), nil
}

func (s PIPSelector) calculate(ctx *Context) (AttributeValue, error) {
	fmt.Printf("s is %v\n", s)

	/*
			item, err := ctx.getContentItem(s.content, s.item)
			fmt.Printf("item is %v\n", item)

			if err != nil {
				return undefinedValue, bindError(err, s.describe())
			}

			if s.t != item.t {
				return undefinedValue, bindError(newInvalidContentItemTypeError(s.t, item.t), s.describe())
			}

		r, err := s.getCategory(s.path, ctx)
		fmt.Printf("r is %v\n", r)

		if err != nil {
			return undefinedValue, bindError(err, s.describe())
		}

		t := r.GetResultType()
		if t != s.t {
			return undefinedValue, bindError(newInvalidContentItemTypeError(s.t, t), s.describe())
		}

		return r, nil
	*/
	return s.getAttributeValue(ctx)
}
