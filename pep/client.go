// Package pep implements gRPC client for Policy Decision Point (PDP) server.
// PEP package (Policy Enforcement Point) wraps service part of golang gRPC
// protocol implementation. The protocol is defined by
// github.com/infobloxopen/themis/proto/service.proto. Its golang implementation
// can be found at github.com/infobloxopen/themis/pdp-service. PEP is able
// to work with single server as well as multiple servers balancing requests
// using round-robin approach.
package pep

//go:generate bash -c "mkdir -p $GOPATH/src/github.com/infobloxopen/themis/pdp-service && protoc -I $GOPATH/src/github.com/infobloxopen/themis/proto/ $GOPATH/src/github.com/infobloxopen/themis/proto/service.proto --go_out=plugins=grpc:$GOPATH/src/github.com/infobloxopen/themis/pdp-service && ls $GOPATH/src/github.com/infobloxopen/themis/pdp-service"

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	ot "github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"

	pb "github.com/infobloxopen/themis/pdp-service"
)

var (
	// ErrorConnected occurs if method connect is called after connection has been established.
	ErrorConnected = errors.New("connection has been already established")
	// ErrorNotConnected indicates that there is no connection established to PDP server.
	ErrorNotConnected = errors.New("no connection")
)

// Client defines abstract PDP service client interface.
//
// Marshalling and unmarshalling
//
// Validate method accepts as "in" argument any structure and pointer to
// any structure as "out" argument. If "in" argument is Request structure from
// github.com/infobloxopen/themis/pdp-service package, Validate passes it as is
// to server. Similarly if "out" argument is pointer to Response structure
// from the protocol package, Validate just copy data from server's response
// to the structure.
//
// If "in" argument is just a structure, Validate marshals it to list of PDP
// attributes. If no fields contains format string, Validate tries to convert
// all exported fields to attributes. Any bool field is converted to boolean
// attribute, string - to string attribute, net.IP - to address, net.IPNet or
// *net.IPNet to network. Fields of other types are silently ingnored.
//
// Marshalling can be ajusted more precisely with help of `pdp` key in format
// string. When some fields of "in" structure have format string, only fields
// with "pdp" key are converted to attributes. The key supports two option
// separated by comma. First is desired attribute name. Second - attribute type.
// Allowed types are: boolean, string, address, network and domain. Validate can
// convert only bool structure field to boolean attribute, string to string
// attribute, net.IP to address attribute, net.IPNet or *net.IPNet to network
// attribute and string to domain attribute.
//
// Validate is also able to unmarshal server's response to structure.
// It accepts pointer to the structure as "out" argument. If no fields
// of the structure has format string, Validate assigns effect to Effect field,
// reason to Reason field and obligation attributes to other fields
// according to their names and types. Effect field can be of bool type
// (and becomes true if effect is Permit or false otherwise), integer (it gets
// one of Response_* constants form pdp-service package) or string (gets
// Response_Effect_name value). Reason should be a string field. Obligation
// attributes are assigned to fields with corresponding names if
// types of fields allow assignment if there is no field with appropriate
// name and type response attribute silently dropped. The same as for marshaling
// `pdp` key can control unmarshaling.
type Client interface {
	// Connect establishes connection to PDP server.
	Connect() error
	// Close terminates previously established connection if any.
	// Close should silently return if connection hasn't been established yet or
	// if it has been already closed.
	Close()

	// Validate sends decision request to PDP server and fills out response.
	Validate(ctx context.Context, in, out interface{}) error
	// ModalValidate is the same as Validate but uses context.Background.
	ModalValidate(in, out interface{}) error
}

type pdpClient struct {
	addr     string
	balancer grpc.Balancer
	conn     *grpc.ClientConn
	client   *pb.PDPClient
	tracer   ot.Tracer
}

// NewBalancedClient creates client instance bound to several PDP servers with round-robin balancing.
func NewBalancedClient(addrs []string, tracer ot.Tracer) Client {
	c := &pdpClient{addr: "pdp", tracer: tracer}
	r := newStaticResolver("pdp", addrs...)
	c.balancer = grpc.RoundRobin(r)
	return c
}

// NewClient creates client instance bound to single PDP server.
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

func makeRequest(v interface{}) (pb.Request, error) {
	if req, ok := v.(pb.Request); ok {
		return req, nil
	}
	attrs, err := marshalValue(reflect.ValueOf(v))
	if err != nil {
		return pb.Request{}, err
	}

	return pb.Request{Attributes: attrs}, nil
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
