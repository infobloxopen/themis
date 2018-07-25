package pipclient

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/infobloxopen/themis/pdp"
	pb "github.com/infobloxopen/themis/pip-service"
)

var allPermitResponse = new(pb.Msg)

func init() {
	b, err := pdp.Response{
		Effect: pdp.EffectPermit,
		Obligations: []pdp.AttributeAssignment{
			pdp.MakeStringAssignment("x", "AllPermitRule"),
		},
	}.Marshal(nil)
	if err != nil {
		panic(err)
	}

	allPermitResponse.Body = b
}

type decisionRequest struct {
	Direction string `pdp:"k1"`
	Policy    string `pdp:"k2"`
	Domain    string `pdp:"k3,domain"`
}

type decisionResponse struct {
	Effect int    `pdp:"Effect"`
	Reason error  `pdp:"Reason"`
	X      string `pdp:"x"`
}

func (r decisionResponse) String() string {
	if r.Reason != nil {
		return fmt.Sprintf("Effect: %q, Reason: %q, X: %q",
			pdp.EffectNameFromEnum(r.Effect),
			r.Reason,
			r.X,
		)
	}

	return fmt.Sprintf("Effect: %q, X: %q", pdp.EffectNameFromEnum(r.Effect), r.X)
}

func TestUnaryClientValidation(t *testing.T) {
	s, err := newAllPermitServer("127.0.0.1:5555")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	t.Run("fixed-buffer", testSingleRequest())
	t.Run("auto-buffer", testSingleRequest(WithAutoRequestSize(true)))
}

func testSingleRequest(opt ...Option) func(t *testing.T) {
	return func(t *testing.T) {
		c := NewClient(opt...)
		err := c.Connect("127.0.0.1:5555")
		if err != nil {
			t.Fatalf("expected no error but got %s", err)
		}
		defer c.Close()

		in := decisionRequest{
			Direction: "Any",
			Policy:    "AllPermitPolicy",
			Domain:    "example.com",
		}
		var out decisionResponse
		err = c.Map(in, &out)
		if err != nil {
			t.Errorf("expected no error but got %s", err)
		}

		if out.Effect != pdp.EffectPermit || out.Reason != nil || out.X != "AllPermitRule" {
			t.Errorf("got unexpected response: %s", out)
		}
	}
}

func TestUnaryClientValidationWithCache(t *testing.T) {
	s, err := newAllPermitServer("127.0.0.1:5555")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	c := NewClient(
		WithMaxRequestSize(128),
		WithCacheTTL(15*time.Minute),
	)
	if err := c.Connect("127.0.0.1:5555"); err != nil {
		t.Fatalf("expected no error but got %s", err)
	}
	defer c.Close()

	uc, ok := c.(*unaryClient)
	if !ok {
		t.Fatalf("expected *unaryClient but got %#v", c)
	}
	bc := uc.cache
	if bc == nil {
		t.Fatal("expected cache")
	}

	in := decisionRequest{
		Direction: "Any",
		Policy:    "AllPermitPolicy",
		Domain:    "example.com",
	}
	var out decisionResponse
	if err := c.Map(in, &out); err != nil {
		t.Errorf("expected no error but got %s", err)
	}

	if out.Effect != pdp.EffectPermit || out.Reason != nil || out.X != "AllPermitRule" {
		t.Errorf("got unexpected response: %s", out)
	}

	if bc.Len() == 1 {
		if it := bc.Iterator(); it.SetNext() {
			ei, err := it.Value()
			if err != nil {
				t.Errorf("can't get value from cache: %s", err)
			} else if err := fillResponse(pb.Msg{Body: ei.Value()}, &out); err != nil {
				t.Errorf("can't unmarshal response from cache: %s", err)
			} else if out.Effect != pdp.EffectPermit || out.Reason != nil || out.X != "AllPermitRule" {
				t.Errorf("got unexpected response from cache: %s", out)
			}
		} else {
			t.Error("can't set cache iterator to the first value")
		}
	} else {
		t.Errorf("expected the only record in cache but got %d", bc.Len())
	}

	if err := c.Map(in, &out); err != nil {
		t.Errorf("expected no error but got %s", err)
	}

	if out.Effect != pdp.EffectPermit || out.Reason != nil || out.X != "AllPermitRule" {
		t.Errorf("got unexpected response: %s", out)
	}
}

type allPermitServer struct {
	s *grpc.Server
}

func newAllPermitServer(addr string) (*allPermitServer, error) {
	if err := waitForPortClosed(addr); err != nil {
		return nil, err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &allPermitServer{s: grpc.NewServer()}
	pb.RegisterPIPServer(s.s, s)
	go s.s.Serve(ln)

	if err := waitForPortOpened(addr); err != nil {
		s.Stop()
		return nil, err
	}

	return s, nil
}

func (s *allPermitServer) Stop() {
	s.s.Stop()
}

func (s *allPermitServer) Map(ctx context.Context, in *pb.Msg) (*pb.Msg, error) {
	return allPermitResponse, nil
}

func (s *allPermitServer) NewMappingStream(stream pb.PIP_NewMappingStreamServer) error {
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		err = stream.Send(allPermitResponse)
		if err != nil {
			return err
		}
	}

	return nil
}

func waitForPortOpened(address string) error {
	var (
		c   net.Conn
		err error
	)

	for i := 0; i < 20; i++ {
		after := time.After(500 * time.Millisecond)
		c, err = net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			return c.Close()
		}

		<-after
	}

	return err
}

func waitForPortClosed(address string) error {
	var (
		c   net.Conn
		err error
	)

	for i := 0; i < 20; i++ {
		after := time.After(500 * time.Millisecond)
		c, err = net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err != nil {
			return nil
		}

		c.Close()
		<-after
	}

	return fmt.Errorf("port at %s hasn't been closed yet", address)
}
