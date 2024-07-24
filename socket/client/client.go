package socketclient

import (
	"net"

	"github.com/willoma/swaypanion/socket"
)

type Client struct {
	conn net.Conn
}

func New() (*Client, error) {
	conn, err := net.Dial("unix", socket.Path)

	return &Client{conn: conn}, err
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Send(msg *socket.Message) error {
	_, err := msg.WriteTo(c.conn)
	return err
}

func (c *Client) Read() (*socket.Message, error) {
	msg := &socket.Message{}
	if _, err := msg.ReadFrom(c.conn); err != nil {
		return nil, err
	}

	return msg, nil
}
