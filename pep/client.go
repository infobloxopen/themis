package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/policy-box/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/policy-box/proto/ $GOPATH/src/github.com/infobloxopen/policy-box/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/policy-box/pdp-service && ls $GOPATH/src/github.com/infobloxopen/policy-box/pdp-service"

import (
	"fmt"
	"reflect"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/infobloxopen/policy-box/pdp-service"
)

var (
	ErrorConnected    = fmt.Errorf("Connection has been already established")
	ErrorNotConnected = fmt.Errorf("No connection")
)

type Client struct {
	addr   string
	conn   *grpc.ClientConn
	client *pb.PDPClient
}

func NewClient(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) Connect(timeout time.Duration) error {
	if c.conn != nil {
		return ErrorConnected
	}

	conn, err := grpc.Dial(c.addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout))
	if err != nil {
		return err
	}

	c.conn = conn

	client := pb.NewPDPClient(c.conn)
	c.client = &client

	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.client = nil
}

func (c *Client) Validate(in, out interface{}) error {
	if c.client == nil {
		return ErrorNotConnected
	}

	req, err := makeRequest(in)
	if err != nil {
		return err
	}

	res, err := (*c.client).Validate(context.Background(), &req)
	if err != nil {
		return err
	}

	return fillResponse(res, out)
}

func makeRequest(v interface{}) (pb.Request, error) {
	if req, ok := v.(pb.Request); ok {
		return req, nil
	}
	attrs, err := marshalValue(reflect.ValueOf(v))
	if err != nil {
		return pb.Request{}, err
	}

	return pb.Request{attrs}, nil
}

func fillResponse(res *pb.Response, v interface{}) error {
	return unmarshalToValue(res, reflect.ValueOf(v))
}
