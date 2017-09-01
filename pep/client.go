package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"fmt"
	"reflect"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"

	pb "github.com/infobloxopen/themis/pdp-service"

	ot "github.com/opentracing/opentracing-go"
)

var (
	ErrorConnected    = fmt.Errorf("Connection has been already established")
	ErrorNotConnected = fmt.Errorf("No connection")
)

type Client interface {
	Connect() error
	Close()

	Validate(ctx context.Context, in, out interface{}) error
	ModalValidate(in, out interface{}) error
}

type pdpClient struct {
	addr     string
	balancer grpc.Balancer
	conn     *grpc.ClientConn
	client   *pb.PDPClient
	tracer   ot.Tracer
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

func (c *pdpClient) ModalValidate(in, out interface{}) error {
	return c.Validate(context.Background(), in, out)
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

func (c *TestClient) ModalValidate(in, out interface{}) error {
	return c.Validate(context.Background(), in, out)
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
	if out, ok := v.(*pb.Response); ok {
		*out = *res
		return nil
	}

	return unmarshalToValue(res, reflect.ValueOf(v))
}

type staticResolver struct {
	Name  string
	Addrs []string
}

func newStaticResolver(name string, addrs ...string) naming.Resolver {
	return &staticResolver{Name: name, Addrs: addrs}
}

func (r *staticResolver) Resolve(target string) (naming.Watcher, error) {
	if target != r.Name {
		return nil, fmt.Errorf("%q is an invalid target for resolver %q", target, r.Name)
	}

	return &staticWatcher{Addrs: r.Addrs}, nil
}

type staticWatcher struct {
	Addrs []string
	stop  chan bool
	sent  bool
}

func (w *staticWatcher) Next() ([]*naming.Update, error) {
	if w.sent {
		stop := <-w.stop
		if stop {
			return nil, nil
		}
	}
	w.stop = make(chan bool)
	w.sent = true
	u := make([]*naming.Update, len(w.Addrs))
	for i, a := range w.Addrs {
		u[i] = &naming.Update{Op: naming.Add, Addr: a}
	}
	return u, nil
}

func (w *staticWatcher) Close() {
	w.stop <- true
}
