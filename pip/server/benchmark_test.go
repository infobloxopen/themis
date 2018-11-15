package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func BenchmarkServer(b *testing.B) {
	s := startBenchEchoServer(b)
	defer s.stop(b)

	reqs := genReqs(128, 16)

	b.Run("BenchmarkServer", func(b *testing.B) {
		a := s.s.ln.Addr()
		c, err := net.Dial(a.Network(), a.String())
		if err != nil {
			b.Fatalf("failed to dial to server %s", err)
		}
		defer c.Close()

		out := make([]byte, msgSizeBytes+msgIdxBytes+31)
		for i := 0; i < b.N; i++ {
			in := reqs[i&127]
			if _, err := c.Write(in); err != nil {
				b.Fatalf("failed to send data %d to %s: %s", i, a, err)
			}

			n, err := c.Read(out)
			if err != nil {
				b.Fatalf("failed to receive data %d from %s: %s", i, a, err)
			}

			if n != len(in) {
				b.Fatalf("got wrong reply %d from %s: %d (n) != %d (len(in))\n\tin.: % x\n\tout: % x",
					i, a, n, len(in), in, out)
			}
		}
	})
}

func BenchmarkServerParallel(b *testing.B) {
	s := startBenchEchoServer(b)
	defer s.stop(b)

	N := 256
	gmp := runtime.GOMAXPROCS(0)
	if N*gmp > math.MaxUint16+1 {
		b.Fatalf("too many goroutines %d > %d", N*gmp, math.MaxUint16+1)
	}

	a := s.s.ln.Addr()
	c := startAsyncBenchClient(b, N*gmp, a)
	defer c.stop()

	p := new(int64)
	b.Run("BenchmarkServerParallel", func(b *testing.B) {
		*p = 0
		b.SetParallelism(N)
		b.RunParallel(func(pb *testing.PB) {
			atomic.AddInt64(p, 1)
			for pb.Next() {
				i, b := c.get()

				r := c.call(i, b)
				if len(r) != len(b) {
					panic(fmt.Errorf("got wrong reply %d from %s: %d len(r) != %d len(b)\n\tr: % x\n\tb: % x",
						i, a, len(r), len(b), r, b))
				}

				c.put(i)
			}
		})
	})

	b.Logf("number of goroutines: %d (GOMAXPROCS=%d)", *p, gmp)
	b.Logf("writes: %.02f rq/chunk %.02f b/chunk", float64(c.wrq)/float64(c.wc), float64(c.wb)/float64(c.wc))
	b.Logf("reads: %.02f re/chunk %.02f b/chunk", float64(c.rre)/float64(c.rc), float64(c.rb)/float64(c.rc))
}

type benchEchoServer struct {
	s   *Server
	wg  *sync.WaitGroup
	err error
}

func startBenchEchoServer(b *testing.B) *benchEchoServer {
	out := &benchEchoServer{
		s: NewServer(
			WithNetwork("tcp4"),
			WithAddress("localhost:0"),
		),
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

const msgIdxBytes = 2

func genReqs(n, m int) [][]byte {
	out := make([][]byte, n)
	for i := range out {
		b := make([]byte, msgSizeBytes+rand.Intn(2*m)+msgIdxBytes)

		binary.LittleEndian.PutUint32(b, uint32(len(b)-msgSizeBytes))
		for j := msgSizeBytes + msgIdxBytes; j < len(b); j++ {
			b[j] = byte(rand.Intn(256))
		}

		out[i] = b
	}

	return out
}

type asyncBenchClient struct {
	reqs [][]byte
	ress [][]byte
	in   chan []byte
	out  []chan []byte
	idx  chan int
	c    net.Conn
	wrq  int
	wc   int
	wb   int
	rre  int
	rc   int
	rb   int
}

func startAsyncBenchClient(b *testing.B, n int, a net.Addr) *asyncBenchClient {
	out := &asyncBenchClient{
		reqs: genReqs(n, 16),
		ress: make([][]byte, n),
		in:   make(chan []byte, n),
		out:  make([]chan []byte, n),
		idx:  make(chan int, n),
	}

	for i := 0; i < n; i++ {
		binary.LittleEndian.PutUint16(out.reqs[i][msgSizeBytes:], uint16(i))
		out.ress[i] = make([]byte, 0, 36)
		out.out[i] = make(chan []byte, 1)
		out.idx <- i
	}

	c, err := net.Dial(a.Network(), a.String())
	if err != nil {
		b.Fatalf("failed to dial to server %s", err)
	}
	out.c = c

	go out.writer()
	go out.reader()

	return out
}

func (c *asyncBenchClient) stop() {
	close(c.in)
	c.c.Close()
}

func (c *asyncBenchClient) get() (int, []byte) {
	i := <-c.idx
	return i, c.reqs[i]
}

func (c *asyncBenchClient) call(i int, in []byte) []byte {
	c.in <- in
	return <-c.out[i]
}

func (c *asyncBenchClient) put(i int) {
	c.idx <- i
}

func (c *asyncBenchClient) reader() {
	buf := make([]byte, defBufSize)
	sBuf := make([]byte, 0, msgSizeBytes)
	iBuf := make([]byte, 0, msgIdxBytes)
	var (
		size int
		mBuf []byte
	)

	i := -1

	for {
		n, err := c.c.Read(buf)
		if n > 0 {
			c.rc++
			c.rb += n
			b := buf[:n]
			for len(b) > 0 {
				if size > 0 {
					if mBuf == nil {
						m := cap(iBuf) - len(iBuf)
						if len(b) < m {
							iBuf = append(iBuf, b...)
							size -= len(b)
							break
						}

						iBuf = append(iBuf, b[:m]...)
						b = b[m:]

						i = int(binary.LittleEndian.Uint16(iBuf))
						mBuf = c.ress[i]

						mBuf = append(mBuf, sBuf...)
						sBuf = sBuf[:0]

						mBuf = append(mBuf, iBuf...)
						iBuf = iBuf[:0]

						size -= m
						if size <= 0 {
							c.out[i] <- mBuf
							i = -1
							mBuf = nil
						}
					} else {
						if len(b) < size {
							mBuf = append(mBuf, b...)
							size -= len(b)
							break
						}

						mBuf = append(mBuf, b[:size]...)
						b = b[size:]
						size = 0

						c.out[i] <- mBuf
						c.rre++
						i = -1
						mBuf = nil
					}
				} else {
					m := cap(sBuf) - len(sBuf)
					if len(b) < m {
						sBuf = append(sBuf, b...)
						break
					}

					sBuf = append(sBuf, b[:m]...)
					b = b[m:]

					size = int(binary.LittleEndian.Uint32(sBuf))
				}
			}
		}

		if err == io.EOF || isConnClosed(err) {
			break
		}

		if err != nil {
			panic(err)
		}
	}
}

func (c *asyncBenchClient) writer() {
	buf := make([]byte, 0, defBufSize)

	t := time.NewTicker(defWriteInt)
	defer t.Stop()
	for {
		select {
		case b, ok := <-c.in:
			if !ok {
				if len(buf) > 0 {
					if _, err := c.c.Write(buf); err != nil {
						panic(err)
					}

					c.wc++
					c.wb += len(buf)
				}

				return
			}

			if cap(buf)-len(buf) < len(b) {
				if _, err := c.c.Write(buf); err != nil {
					panic(err)
				}

				c.wc++
				c.wb += len(buf)
				buf = buf[:0]
			}

			buf = append(buf, b...)
			c.wrq++

		case <-t.C:
			if len(buf) > 0 {
				if _, err := c.c.Write(buf); err != nil {
					panic(err)
				}

				c.wc++
				c.wb += len(buf)
				buf = buf[:0]
			}
		}
	}
}
