package testutil

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/infobloxopen/themis/pdp"
	_ "github.com/infobloxopen/themis/pdp/selector"
	"github.com/infobloxopen/themis/pdpserver/server"
	"github.com/infobloxopen/themis/pep"
	logr "github.com/sirupsen/logrus"
)

func StartPDPServer(t *testing.T, p, endpoint string) *LoggedServer {
	s := NewServer(server.WithServiceAt(endpoint))

	if err := s.S.ReadPolicies(strings.NewReader(p)); err != nil {
		t.Fatalf("can't read policies: %s", err)
	}

	if err := WaitForPortClosed(endpoint); err != nil {
		t.Fatalf("port still in use: %s", err)
	}

	go func() {
		if err := s.S.Serve(); err != nil {
			t.Fatalf("PDP server failed: %s", err)
		}
	}()

	return s
}

type LoggedServer struct {
	S *server.Server
	B *bytes.Buffer
}

func NewServer(opts ...server.Option) *LoggedServer {
	s := &LoggedServer{
		B: new(bytes.Buffer),
	}

	logger := logr.New()
	logger.Out = s.B
	logger.Level = logr.ErrorLevel
	opts = append(opts,
		server.WithLogger(logger),
	)

	s.S = server.NewServer(opts...)
	return s
}

func (s *LoggedServer) Stop() string {
	s.S.Stop()
	return s.B.String()
}

func WaitForPortOpened(address string) error {
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

func WaitForPortClosed(address string) error {
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

type logGrabber struct {
	b *bytes.Buffer
}

func NewLogGrabber() *logGrabber {
	b := new(bytes.Buffer)
	logr.SetOutput(b)

	return &logGrabber{
		b: b,
	}
}

func (g *logGrabber) Release() string {
	logr.SetOutput(os.Stderr)

	return g.b.String()
}

type erraticPep struct {
	counter int
	err     []error
	client  pep.Client
}

func NewErraticPep(c pep.Client, err ...error) *erraticPep {
	return &erraticPep{
		err:    err,
		client: c,
	}
}

func (c *erraticPep) Connect(addr string) error {
	return c.client.Connect(addr)
}

func (c *erraticPep) Close() {
	c.client.Close()
}

func (c *erraticPep) Validate(in, out interface{}) error {
	if len(c.err) > 0 {
		n := c.counter % len(c.err)
		c.counter++

		err := c.err[n]
		if err != nil {
			return err
		}
	}

	return c.client.Validate(in, out)
}

type MockPdpClient struct {
	T      *testing.T
	In     []pdp.AttributeAssignment
	Out    []pdp.AttributeAssignment
	Effect int
	Status error
	Err    error
}

func NewMockPdpClient(t *testing.T) *MockPdpClient {
	return &MockPdpClient{T: t}
}

func (mpc *MockPdpClient) Connect(addr string) error {
	return nil
}

func (mpc *MockPdpClient) Close() {}

func (mpc *MockPdpClient) Validate(in, out interface{}) error {
	if mpc.In != nil {
		inAttrs, ok := in.([]pdp.AttributeAssignment)
		if !ok {
			mpc.T.Errorf("Incorrect input data type")
		}
		AssertAttrList(mpc.T, inAttrs, mpc.In...)
	}
	resp, ok := out.(*pdp.Response)
	if !ok {
		mpc.T.Fatal("Incorrect output data type")
	}
	resp.Effect = mpc.Effect
	resp.Obligations = mpc.Out
	resp.Status = mpc.Status
	return mpc.Err
}
