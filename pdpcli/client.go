package main

import (
	"fmt"

	"google.golang.org/grpc"
)

var (
    ErrorConnected    = fmt.Errorf("Connection has been already established")
    ErrorNotConnected = fmt.Errorf("No connection")
)

type Client struct {
	Address string

	conn *grpc.ClientConn
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Connect(addr string) error {
	if c.conn != nil {
		return ErrorConnected
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	c.Address = addr
	c.conn = conn
	return nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
