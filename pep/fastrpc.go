package pep

import (
	"math/rand"
	"net"
	"sync"
	"time"

	pdp "github.com/infobloxopen/themis/pdp-service"
	"github.com/valyala/fastrpc"
	"github.com/valyala/fastrpc/tlv"
)

const (
	deadline = 1 * time.Second
)

type client struct {
	endpoints          []string
	maxBatchDelay      int
	maxPendingRequests int
	clients            []*fastrpc.Client
	conns              []net.Conn
	sync.Mutex
}

type Client interface {
	Connect() error
	Close()
	Validate(request *pdp.Request) (*pdp.Response, error)
}

func NewClient(endpoints []string, delay, pending uint) Client {
	if pending <= 0 {
		pending = fastrpc.DefaultMaxPendingRequests
	}
	if delay <= 0 {
		delay = 100
	}
	return &client{
		endpoints:          endpoints,
		maxBatchDelay:      int(delay),
		maxPendingRequests: int(pending),
		clients:            make([]*fastrpc.Client, len(endpoints)),
	}
}

func (c *client) dial(addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	c.Lock()
	defer c.Unlock()
	if err == nil {
		c.conns = append(c.conns, conn)
	}
	return conn, err
}

func (c *client) Connect() error {
	for i, endpoint := range c.endpoints {
		c.clients[i] = &fastrpc.Client{
			Addr:               endpoint,
			CompressType:       fastrpc.CompressNone,
			MaxBatchDelay:      time.Duration(c.maxBatchDelay) * time.Microsecond,
			MaxPendingRequests: c.maxPendingRequests,
			NewResponse: func() fastrpc.ResponseReader {
				return &tlv.Response{}
			},
			Dial: c.dial,
		}
	}
	return nil
}

func (c *client) Close() {
	c.Lock()
	defer c.Unlock()
	for _, conn := range c.conns {
		conn.Close()
	}
}

func (c *client) client() *fastrpc.Client {
	l := len(c.clients)
	index := 0
	if l > 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		index = r.Intn(l)
	}
	return c.clients[index]
}

func (c *client) Validate(request *pdp.Request) (*pdp.Response, error) {
	var req tlv.Request
	var resp tlv.Response
	req.SwapValue(pdp.MarshalRequest(request))
	err := c.client().DoDeadline(&req, &resp, time.Now().Add(deadline))
	if err != nil {
		return nil, err
	}
	ret := pdp.UnmarshalResponse(resp.Value())
	return ret, nil
}
