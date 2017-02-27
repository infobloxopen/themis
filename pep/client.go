package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/policy-box/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/policy-box/proto/ $GOPATH/src/github.com/infobloxopen/policy-box/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/policy-box/pdp-service && ls $GOPATH/src/github.com/infobloxopen/policy-box/pdp-service"

import (
	"fmt"
	"reflect"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"

	pb "github.com/infobloxopen/policy-box/pdp-service"

	ot "github.com/opentracing/opentracing-go"
)

var (
	ErrorConnected    = fmt.Errorf("Connection has been already established")
	ErrorNotConnected = fmt.Errorf("No connection")
)

type Client struct {
	addr   string
	conn   *grpc.ClientConn
	client *pb.PDPClient
	tracer ot.Tracer
}

func NewClient(addr string, tracer ot.Tracer) *Client {
	return &Client{addr: addr, tracer: tracer}
}

func (c *Client) Connect(timeout time.Duration) error {
	if c.conn != nil {
		return ErrorConnected
	}
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(timeout))
	if c.tracer != nil {
		onlyIfParent := func(parentSpanCtx ot.SpanContext, method string, req, resp interface{}) bool {
			return parentSpanCtx != nil
		}
		intercept := otgrpc.OpenTracingClientInterceptor(c.tracer, otgrpc.IncludingSpans(onlyIfParent))
		dialOpts = append(dialOpts, grpc.WithUnaryInterceptor(intercept))
	}
	conn, err := grpc.Dial(c.addr, dialOpts...)
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

func (c *Client) Validate(ctx context.Context, in, out interface{}) error {
	if c.client == nil {
		return ErrorNotConnected
	}

	req, err := makeRequest(in)
	if err != nil {
		return err
	}

	res, err := (*c.client).Validate(ctx, &req)
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
