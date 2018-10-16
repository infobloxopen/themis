package client

import (
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

	a := pdp.MakeStringAssignment("test", "test")

	b.Run("BenchmarkSequential", func(b *testing.B) {
		c := NewClient()
		if err := c.Connect(); err != nil {
			b.Fatalf("failed to connect to server %s", err)
		}
		defer c.Close()

		for i := 0; i < b.N; i++ {
			if _, err := c.Get(a); err != nil {
				b.Fatalf("failed to get data %d: %s", i, err)
			}
		}
	})
}

func BenchmarkParallel(b *testing.B) {
	s := startBenchEchoServer(b)
	defer s.stop(b)

	a := pdp.MakeStringAssignment("test", "test")

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
				if _, err := c.Get(a); err != nil {
					panic(fmt.Errorf("failed to get data: %s", err))
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

	a := pdp.MakeStringAssignment("test", "test")

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
				if _, err := c.Get(a); err != nil {
					panic(fmt.Errorf("failed to get data: %s", err))
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

	a := pdp.MakeStringAssignment("test", "test")

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
				if _, err := c.Get(a); err != nil {
					panic(fmt.Errorf("failed to get data: %s", err))
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

func startBenchEchoServer(b *testing.B, opts ...server.Option) *benchEchoServer {
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
