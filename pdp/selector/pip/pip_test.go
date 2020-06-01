package pip

import (
	"encoding/binary"
	"fmt"
	"net/url"
	"sync"
	"testing"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pip/server"
)

func TestMakePipSelector(t *testing.T) {
	pc := NewTCPClientsPool()
	e, err := MakePipSelector(
		pc,
		makeTestURL("pip://localhost:5600/content/item"),
		[]pdp.Expression{pdp.MakeStringValue("test")},
		pdp.TypeString,
	)
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	}
	if e == nil {
		t.Errorf("expected PipSelector but got nothing")
	} else if s, ok := e.(PipSelector); !ok {
		t.Errorf("expected PipSelector but got %T (%#v)", e, e)

		if s.clients != pc {
			t.Errorf("expected %#v clients but got %#v", pc, s.clients)
		}

		if s.net != "tcp" {
			t.Errorf("expected %q network but got %q", "tcp", s.net)
		}

		if s.k8s {
			t.Error("expected no kubernetes discovery")
		}

		if s.addr != "localhost:5600" {
			t.Errorf("expected %q address but got %q", "localhost:5600", s.addr)
		}

		if s.id != "content/item" {
			t.Errorf("expected %q as content id but got %q", "content/item", s.id)
		}

		if s.t != pdp.TypeString {
			t.Errorf("exepcted %q type but got %q", pdp.TypeString, s.t)
		}
	}

	uc := NewUnixClientsPool()
	e, err = MakePipSelector(
		uc,
		makeTestURL("pip+unix:/var/run/pip.socket#content/item"),
		[]pdp.Expression{pdp.MakeStringValue("test")},
		pdp.TypeString,
	)
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	}
	if e == nil {
		t.Errorf("expected PipSelector but got nothing")
	} else if s, ok := e.(PipSelector); !ok {
		t.Errorf("expected PipSelector but got %T (%#v)", e, e)

		if s.clients != uc {
			t.Errorf("expected %#v clients but got %#v", uc, s.clients)
		}

		if s.net != "unix" {
			t.Errorf("expected %q network but got %q", "unix", s.net)
		}

		if s.k8s {
			t.Error("expected no kubernetes discovery")
		}

		if s.addr != "/var/run/pip.socket" {
			t.Errorf("expected %q address but got %q", "/var/run/pip.socket", s.addr)
		}

		if s.id != "content/item" {
			t.Errorf("expected %q as content id but got %q", "content/item", s.id)
		}

		if s.t != pdp.TypeString {
			t.Errorf("exepcted %q type but got %q", pdp.TypeString, s.t)
		}
	}

	kc := NewK8sClientsPool()
	e, err = MakePipSelector(
		kc,
		makeTestURL("pip+k8s://value.key.namespace:5600/content/item"),
		[]pdp.Expression{pdp.MakeStringValue("test")},
		pdp.TypeString,
	)
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	}
	if e == nil {
		t.Errorf("expected PipSelector but got nothing")
	} else if s, ok := e.(PipSelector); !ok {
		t.Errorf("expected PipSelector but got %T (%#v)", e, e)

		if s.clients != kc {
			t.Errorf("expected %#v clients but got %#v", kc, s.clients)
		}

		if s.net != "tcp" {
			t.Errorf("expected %q network but got %q", "tcp", s.net)
		}

		if !s.k8s {
			t.Error("expected kubernetes discovery")
		}

		if s.addr != "value.key.namespace:5600" {
			t.Errorf("expected %q address but got %q", "value.key.namespace:5600", s.addr)
		}

		if s.id != "content/item" {
			t.Errorf("expected %q as content id but got %q", "content/item", s.id)
		}

		if s.t != pdp.TypeString {
			t.Errorf("exepcted %q type but got %q", pdp.TypeString, s.t)
		}
	}

	_, err = MakePipSelector(
		pc,
		makeTestURL("local:content/item"),
		[]pdp.Expression{pdp.MakeStringValue("test")},
		pdp.TypeString,
	)
	if err == nil {
		t.Error("expected error")
	}
}

func TestPipSelectorCalculate(t *testing.T) {
	s := startTestEchoServer(t)
	defer s.stop(t)

	done := make(chan struct{})
	defer close(done)

	pc := NewTCPClientsPool()
	go pc.cleaner(nil, done)

	e, err := MakePipSelector(
		pc,
		makeTestURL("pip://localhost:5600/content/item"),
		[]pdp.Expression{pdp.MakeStringValue("test")},
		pdp.TypeString,
	)
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	} else {
		v, err := e.Calculate(nil)
		if err != nil {
			t.Errorf("expected no error but got %#v", err)
		}

		s, err := v.Serialize()
		if err != nil {
			t.Errorf("failed to serialize result %#v", err)
		} else if s != "test" {
			t.Errorf("expected %q from PIP but got %q", "test", s)
		}
	}

	e, err = MakePipSelector(
		pc,
		makeTestURL("pip://localhost:5600/content/item"),
		[]pdp.Expression{pdp.MakeStringValue("test")},
		pdp.TypeInteger,
		pdp.SelectorOption{Name: pdp.SelectorOptionError, Data: pdp.MakeIntegerValue(5)},
	)
	if err != nil {
		t.Errorf("expected no error but got %#v", err)
	} else {
		v, err := e.Calculate(nil)
		if err != nil {
			t.Errorf("expected no error but got %#v", err)
		}

		s, err := v.Serialize()
		if err != nil {
			t.Errorf("failed to serialize result %#v", err)
		} else if s != "5" {
			t.Errorf("expected %q from PIP but got %q", "5", s)
		}
	}
}

func TestPanicOnBadDefaultOption(t *testing.T) {
	checkPanicOnBadOption(t, pdp.SelectorOption{
		Name: pdp.SelectorOptionDefault,
		Data: "must be expression",
	})
}

func TestPanicOnBadErrorOption(t *testing.T) {
	checkPanicOnBadOption(t, pdp.SelectorOption{
		Name: pdp.SelectorOptionError,
		Data: "must be expression",
	})
}

func checkPanicOnBadOption(t *testing.T, opt pdp.SelectorOption) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("no panic on incorrect selector option")
		}
	}()

	MakePipSelector(
		NewTCPClientsPool(),
		makeTestURL("pip://localhost:5600/content/item"),
		[]pdp.Expression{},
		pdp.TypeString,
		opt,
	)
}

func makeTestURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return u
}

type testEchoServer struct {
	s   *server.Server
	wg  *sync.WaitGroup
	err error
}

var testEchoServerInfoResponseHeader = []byte{1, 0, 0, 0, 2}

func startTestEchoServer(t *testing.T, opts ...server.Option) *testEchoServer {
	opts = append(opts,
		server.WithHandler(testEchoServerHandler),
	)

	out := &testEchoServer{
		s:  server.NewServer(opts...),
		wg: new(sync.WaitGroup),
	}

	if err := out.s.Bind(); err != nil {
		t.Fatalf("failed to bind server: %s", err)
	}

	out.wg.Add(1)
	go func() {
		defer out.wg.Done()
		out.err = out.s.Serve()
	}()

	return out
}

func (s *testEchoServer) stop(t *testing.T) {
	if err := s.s.Stop(); err != nil {
		t.Fatalf("failed to stop server: %s", err)
	}

	s.wg.Wait()
	if s.err != nil && s.err != server.ErrNotBound {
		t.Fatalf("failed to start server: %s", s.err)
	}
}

func testEchoServerHandler(b []byte) []byte {
	if len(b) < 4 {
		panic("too short input buffer")
	}

	in := testEchoServerHandlerCheckVersion(b[4:])
	in = testEchoServerHandlerSkipPath(in)
	in = testEchoServerHandlerCheckValuesNumber(in)
	in = testEchoServerHandlerCheckFirstValueType(in)
	in = testEchoServerHandlerGetFirstValueBytes(in)

	out := b
	m := 4
	m += copy(out[m:], testEchoServerInfoResponseHeader)
	binary.LittleEndian.PutUint16(out[m:], uint16(len(in)))
	m += 2
	m += copy(out[m:], in)

	return out[:m]
}

func testEchoServerHandlerCheckVersion(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	if v := binary.LittleEndian.Uint16(b); v != 1 {
		panic(fmt.Errorf("invalid information request version %d (expected %d)", v, 1))
	}

	return b[2:]
}

func testEchoServerHandlerSkipPath(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	n := int(binary.LittleEndian.Uint16(b))
	b = b[2:]

	if len(b) < n {
		panic("too short input buffer")
	}

	return b[n:]
}

func testEchoServerHandlerCheckValuesNumber(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	if c := int(binary.LittleEndian.Uint16(b)); c < 1 {
		panic(fmt.Errorf("expected at least one value but got %d", c))
	}

	return b[2:]
}

func testEchoServerHandlerCheckFirstValueType(b []byte) []byte {
	if len(b) < 1 {
		panic("too short input buffer")
	}

	if t := b[0]; t != 2 {
		panic(fmt.Errorf("expected value of type %d (string) but got %d", 2, t))
	}

	return b[1:]
}

func testEchoServerHandlerGetFirstValueBytes(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	n := int(binary.LittleEndian.Uint16(b))
	b = b[2:]

	if len(b) < n {
		panic(fmt.Errorf("too short input buffer %d < %d", len(b), n))
	}

	return b[:n]
}
