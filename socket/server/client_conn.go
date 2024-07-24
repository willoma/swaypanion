package socketserver

import (
	"errors"
	"io"
	"net"
	"strconv"
	"sync"

	"github.com/willoma/swaypanion/common"
	"github.com/willoma/swaypanion/socket"
)

type Connection struct {
	mu sync.Mutex

	conn io.ReadWriteCloser
}

func (c *Connection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.conn.Close()

}

func (c *Connection) Receive() *socket.Message {
	msg := &socket.Message{}
	if _, err := msg.ReadFrom(c.conn); err != nil {
		if !(errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed)) {
			common.LogError("Failed to receive command from socket client", err)
		}

		return nil
	}

	return msg
}

func (c *Connection) Send(msg socket.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, err := msg.WriteTo(c.conn)

	return err
}

func (c *Connection) SendString(command, value string, complement ...string) error {
	return c.Send(socket.Message{
		Command:    command,
		Value:      value,
		Complement: complement,
	})
}

func (c *Connection) SendError(e string, complements ...string) {
	if err := c.Send(socket.Message{
		Command:    "error",
		Value:      e,
		Complement: complements,
	}); err != nil {
		common.LogError("Failed to send error", err)
	}
}

func (c *Connection) SendInt(command string, value int) error {
	return c.Send(socket.Message{
		Command: command,
		Value:   strconv.Itoa(value),
	})
}
