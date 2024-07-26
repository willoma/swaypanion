package main

import (
	"errors"
	"io"
	"os"

	"github.com/willoma/swaypanion/socket"
)

func (c *client) direct() {
	msg := &socket.Message{
		Command: os.Args[1],
	}

	if len(os.Args) > 2 {
		msg.Value = os.Args[2]
	}

	if len(os.Args) > 3 {
		msg.Complement = os.Args[3:]
	}

	if err := c.client.Send(msg); err != nil {
		c.directPrintError("Failed to send command", err)
	}

	if err := c.client.Send(&socket.Message{
		Command: "close",
	}); err != nil {
		c.directPrintError("Failed to send close command", err)
	}

	for {
		msg, err := c.client.Read()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				c.directPrintError("Failed to read response", err)
			}

			return
		}

		c.directPrintResponse(msg)
	}
}
