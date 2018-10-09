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

func BenchmarkClientServer(b *testing.B) {
	s := startBenchEchoServer(b)
	defer s.stop(b)

	a := pdp.MakeStringAssignment("test", "test")

	b.Run("BenchmarkClientServer", func(b *testing.B) {
		c := NewClient()
		if err := c.Connect(); err != nil {
			b.Fatalf("failed to connect to server %s", err)
		}
		defer c.Close()

		for i := 0; i < b.N; i++ {
			if _, err := c.Get(a); err != nil {
				n := c.(*client).b.(*simpleBalancer).c.n
				b.Fatalf("failed to get data %d from %s: %s", i, n.RemoteAddr(), err)
			}
		}
	})
}

func BenchmarkClientServerParallel(b *testing.B) {
	s := startBenchEchoServer(b)
	defer s.stop(b)

	a := pdp.MakeStringAssignment("test", "test")

	N := 256
	gmp := runtime.GOMAXPROCS(0)

	c := NewClient(WithMaxQueue(N * gmp))
	if err := c.Connect(); err != nil {
		b.Fatalf("failed to connect to server %s", err)
	}
	defer c.Close()

	p := new(int64)
	idx := 0
	b.Run("BenchmarkClientServerParallel", func(b *testing.B) {
		idx++
		*p = 0
		b.SetParallelism(N)
		b.RunParallel(func(pb *testing.PB) {
			atomic.AddInt64(p, 1)
			for pb.Next() {
				if _, err := c.Get(a); err != nil {
					n := c.(*client).b.(*simpleBalancer).c.n
					panic(fmt.Errorf("failed to get data from %s: %s", n.RemoteAddr(), err))
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

func startBenchEchoServer(b *testing.B) *benchEchoServer {
	out := &benchEchoServer{
		s:  server.NewServer(),
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
	if s.err != nil {
		b.Fatalf("failed to start server: %s", s.err)
	}
}
