package main

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/willoma/swaypanion/socket"
)

func (c *client) direct() {
	msg, err := readMessage(
		strings.NewReader(strings.Join(os.Args[1:], " ")),
	)

	if err != nil && !errors.Is(err, io.EOF) {
		c.directPrintError("Failed to read command", err)
	}

	if err := c.client.Send(msg); err != nil {
		c.directPrintError("Failed to send command", err)
	}

	if err := c.client.Send(&socket.Message{Command: "close"}); err != nil {
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
