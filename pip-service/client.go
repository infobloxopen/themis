package pip_service

import (
	"fmt"

	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Client API for PDP service
type PIPClient interface {
	GetAttribute(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type pIPClient struct {
	cc *grpc.ClientConn
}

func NewPIPClient(cc *grpc.ClientConn) *pIPClient {
	return &pIPClient{cc}
}

var (
	ErrorConnected    = fmt.Errorf("Connection has been already established")
	ErrorNotConnected = fmt.Errorf("No connection")
)

type Client interface {
	Connect() error
	Close()
	Validate(ctx context.Context, in, out interface{}) error
}

type pipClient struct {
	conn     *grpc.ClientConn
	client   *pb.PIPClient
}

func NewBalancedClient(addrs []string, tracer ot.Tracer) Client {
	c := &pdpClient{addr: "pdp", tracer: tracer}
	r := newStaticResolver("pdp", addrs...)
	c.balancer = grpc.RoundRobin(r)
	return c
}

func NewClient(addr string, tracer ot.Tracer) Client {
	return &pdpClient{addr: addr, tracer: tracer}
}

func (c *pdpClient) Connect() error {
	if c.conn != nil {
		return ErrorConnected
	}
	var dialOpts []grpc.DialOption
	dialOpts = append(dialOpts, grpc.WithInsecure())
	if c.balancer != nil {
		dialOpts = append(dialOpts, grpc.WithBalancer(c.balancer))
	}
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

func (c *pdpClient) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	c.client = nil
}

func (c *pdpClient) Validate(ctx context.Context, in, out interface{}) error {
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

type TestClient struct {
	NextResponse   *pb.Response
	NextResponseIP *pb.Response
	ErrResponse    error
	ErrResponseIP  error
}

func NewTestClient() *TestClient {
	return &TestClient{}
}

func NewTestClientInit(nextResponse *pb.Response, nextResponseIP *pb.Response,
	errResponse error, errResponseIP error) *TestClient {
	return &TestClient{nextResponse, nextResponseIP, errResponse, errResponseIP}
}

func (c *TestClient) Connect() error { return nil }
func (c *TestClient) Close()         {}
func (c *TestClient) Validate(ctx context.Context, in, out interface{}) error {
	if in != nil {
		p := in.(pb.Request)
		for _, a := range p.Attributes {
			if a.Id == "address" {
				if c.ErrResponseIP != nil {
					return c.ErrResponseIP
				}
				if c.NextResponseIP != nil {
					return fillResponse(c.NextResponseIP, out)
				}
				continue
			}
		}
	}
	if c.ErrResponse != nil {
		return c.ErrResponse
	}
	return fillResponse(c.NextResponse, out)
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