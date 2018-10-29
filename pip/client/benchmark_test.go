package client

import (
	"encoding/binary"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/infobloxopen/themis/pdp"
	"github.com/infobloxopen/themis/pip/server"
)

const N = 256

func BenchmarkSequential(b *testing.B) {
	s := startBenchEchoServer(b)
	defer s.stop(b)

	a := []pdp.AttributeValue{
		pdp.MakeStringValue("test"),
	}

	b.Run("BenchmarkSequential", func(b *testing.B) {
		c := NewClient()
		if err := c.Connect(); err != nil {
			b.Fatalf("failed to connect to server %s", err)
		}
		defer c.Close()

		for i := 0; i < b.N; i++ {
			v, err := c.Get("test", a)
			if err != nil {
				b.Fatalf("failed to get data %d: %s", i, err)
			}

			s, err := v.Serialize()
			if err != nil {
				b.Fatalf("failed to serialize returned value %d: %s", i, err)
			}

			if s != "test" {
				b.Fatalf("expected %q for %d but got %q", "test", i, s)
			}
		}
	})
}

func BenchmarkParallel(b *testing.B) {
	s := startBenchEchoServer(b)
	defer s.stop(b)

	a := []pdp.AttributeValue{
		pdp.MakeStringValue("test"),
	}

	gmp := runtime.GOMAXPROCS(0)
	c := NewClient(WithMaxQueue(N * gmp))
	if err := c.Connect(); err != nil {
		b.Fatalf("failed to connect to server %s", err)
	}
	defer c.Close()

	p := new(int64)
	idx := 0

	b.Run("BenchmarkParallel", func(b *testing.B) {
		idx++
		*p = 0
		b.SetParallelism(N)
		b.RunParallel(func(pb *testing.PB) {
			atomic.AddInt64(p, 1)
			for pb.Next() {
				v, err := c.Get("test", a)
				if err != nil {
					panic(fmt.Errorf("failed to get data: %s", err))
				}

				s, err := v.Serialize()
				if err != nil {
					panic(fmt.Errorf("failed to serialize returned value: %s", err))
				}

				if s != "test" {
					panic(fmt.Errorf("expected %q but got %q", "test", s))
				}
			}
		})
	})

	b.Logf("number of goroutines: %d (GOMAXPROCS=%d)", *p, gmp)
}

func BenchmarkRoundRobin(b *testing.B) {
	s1 := startBenchEchoServer(b, server.WithAddress("127.0.0.1:5601"))
	defer s1.stop(b)

	s2 := startBenchEchoServer(b, server.WithAddress("127.0.0.1:5602"))
	defer s2.stop(b)

	s3 := startBenchEchoServer(b, server.WithAddress("127.0.0.1:5603"))
	defer s3.stop(b)

	s4 := startBenchEchoServer(b, server.WithAddress("127.0.0.1:5604"))
	defer s4.stop(b)

	a := []pdp.AttributeValue{
		pdp.MakeStringValue("test"),
	}

	gmp := runtime.GOMAXPROCS(0)
	c := NewClient(
		WithMaxQueue(N*gmp/4),
		WithRoundRobinBalancer(
			"127.0.0.1:5601",
			"127.0.0.1:5602",
			"127.0.0.1:5603",
			"127.0.0.1:5604",
		),
	)
	if err := c.Connect(); err != nil {
		b.Fatalf("failed to connect to server %s", err)
	}
	defer c.Close()

	p := new(int64)
	idx := 0
	b.Run("BenchmarkRoundRobin", func(b *testing.B) {
		idx++
		*p = 0
		b.SetParallelism(N)
		b.RunParallel(func(pb *testing.PB) {
			atomic.AddInt64(p, 1)
			for pb.Next() {
				v, err := c.Get("test", a)
				if err != nil {
					panic(fmt.Errorf("failed to get data: %s", err))
				}

				s, err := v.Serialize()
				if err != nil {
					panic(fmt.Errorf("failed to serialize returned value: %s", err))
				}

				if s != "test" {
					panic(fmt.Errorf("expected %q but got %q", "test", s))
				}
			}
		})
	})

	b.Logf("number of goroutines: %d (GOMAXPROCS=%d)", *p, gmp)
}

func BenchmarkHotSpot(b *testing.B) {
	s1 := startBenchEchoServer(b, server.WithAddress("127.0.0.1:5601"))
	defer s1.stop(b)

	s2 := startBenchEchoServer(b, server.WithAddress("127.0.0.1:5602"))
	defer s2.stop(b)

	s3 := startBenchEchoServer(b, server.WithAddress("127.0.0.1:5603"))
	defer s3.stop(b)

	s4 := startBenchEchoServer(b, server.WithAddress("127.0.0.1:5604"))
	defer s4.stop(b)

	a := []pdp.AttributeValue{
		pdp.MakeStringValue("test"),
	}

	gmp := runtime.GOMAXPROCS(0)
	c := NewClient(
		WithMaxQueue(N*gmp/4),
		WithHotSpotBalancer(
			"127.0.0.1:5601",
			"127.0.0.1:5602",
			"127.0.0.1:5603",
			"127.0.0.1:5604",
		),
	)
	if err := c.Connect(); err != nil {
		b.Fatalf("failed to connect to server %s", err)
	}
	defer c.Close()

	p := new(int64)
	idx := 0
	b.Run("BenchmarkHotSpot", func(b *testing.B) {
		idx++
		*p = 0
		b.SetParallelism(N)
		b.RunParallel(func(pb *testing.PB) {
			atomic.AddInt64(p, 1)
			for pb.Next() {
				v, err := c.Get("test", a)
				if err != nil {
					panic(fmt.Errorf("failed to get data: %s", err))
				}

				s, err := v.Serialize()
				if err != nil {
					panic(fmt.Errorf("failed to serialize returned value: %s", err))
				}

				if s != "test" {
					panic(fmt.Errorf("expected %q but got %q", "test", s))
				}
			}
		})
	})

	b.Logf("number of goroutines: %d (GOMAXPROCS=%d)", *p, gmp)
}

type benchEchoServer struct {
	s   *server.Server
	wg  *sync.WaitGroup
	err error
}

var benchEchoServerInfoResponseHeader = []byte{1, 0, 0, 0, 2}

func startBenchEchoServer(b *testing.B, opts ...server.Option) *benchEchoServer {
	opts = append(opts,
		server.WithHandler(benchEchoServerHandler),
	)

	out := &benchEchoServer{
		s:  server.NewServer(opts...),
		wg: new(sync.WaitGroup),
	}

	if err := out.s.Bind(); err != nil {
		b.Fatalf("failed to bind server: %s", err)
	}

	out.wg.Add(1)
	go func() {
		defer out.wg.Done()
		out.err = out.s.Serve()
	}()

	return out
}

func (s *benchEchoServer) stop(b *testing.B) {
	if err := s.s.Stop(); err != nil {
		b.Fatalf("failed to stop server: %s", err)
	}

	s.wg.Wait()
	if s.err != nil && s.err != server.ErrNotBound {
		b.Fatalf("failed to start server: %s", s.err)
	}
}
func benchEchoServerHandler(b []byte) []byte {
	if len(b) < 4 {
		panic("too short input buffer")
	}

	in := benchEchoServerHandlerCheckVersion(b[4:])
	in = benchEchoServerHandlerSkipPath(in)
	in = benchEchoServerHandlerCheckValuesNumber(in)
	in = benchEchoServerHandlerCheckFirstValueType(in)
	in = benchEchoServerHandlerGetFirstValueBytes(in)

	out := b
	m := 4
	m += copy(out[m:], benchEchoServerInfoResponseHeader)
	binary.LittleEndian.PutUint16(out[m:], uint16(len(in)))
	m += 2
	m += copy(out[m:], in)

	return out[:m]
}

func benchEchoServerHandlerCheckVersion(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	if v := binary.LittleEndian.Uint16(b); v != 1 {
		panic(fmt.Errorf("invalid information request version %d (expected %d)", v, 1))
	}

	return b[2:]
}

func benchEchoServerHandlerSkipPath(b []byte) []byte {
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

func benchEchoServerHandlerCheckValuesNumber(b []byte) []byte {
	if len(b) < 2 {
		panic("too short input buffer")
	}

	if c := int(binary.LittleEndian.Uint16(b)); c < 1 {
		panic(fmt.Errorf("expected at least one value but got %d", c))
	}

	return b[2:]
}

func benchEchoServerHandlerCheckFirstValueType(b []byte) []byte {
	if len(b) < 1 {
		panic("too short input buffer")
	}

	if t := b[0]; t != 2 {
		panic(fmt.Errorf("expected value of type %d (string) but got %d", 2, t))
	}

	return b[1:]
}

func benchEchoServerHandlerGetFirstValueBytes(b []byte) []byte {
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
