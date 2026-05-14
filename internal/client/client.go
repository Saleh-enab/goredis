package client

import "net"

type Client struct {
	Conn          net.Conn
	Authenticated bool
}

func NewClient(conn net.Conn) *Client {
	return &Client{Conn: conn, Authenticated: false}
}
