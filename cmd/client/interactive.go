package main

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/signal"

	"github.com/willoma/swaypanion/socket"
)

const (
	interactiveFieldSeparator byte = ':'
	interactiveEndOfCommand   byte = '\n'
)

func (c *client) interactive() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	go c.readIncomingMessages()
	go c.handleCLI()

	select {
	case <-ctx.Done():
	case <-c.stopped:
	}
}
func (c *client) handleCLI() {
	for {
		msg, err := c.prompt()
		if err != nil {
			if errors.Is(err, io.EOF) {
				c.interactivePrintNewline()
				close(c.stopped)
				return
			}

			c.interactivePrintError("Failed to get instruction", err)
		}

		switch msg.Command {
		case "q", "qu", "qui", "quit", "e", "ex", "exi", "exit":
			close(c.stopped)
			return
		case "help":
			c.interactivePrintLine("exit: Close the swaypanion client")
			c.interactivePrintLine("quit: Close the swaypanion client")
		}

		if err := c.client.Send(msg); err != nil {
			c.interactivePrintError("Failed to send command", err)
		}
	}
}

func (c *client) prompt() (*socket.Message, error) {
	c.interactivePrintPrompt()

	var (
		in           = make([]byte, 1)
		current      []byte
		currentField int
		message      = &socket.Message{}
	)

	assignCurrent := func() {
		switch currentField {
		case 0:
			message.Command = string(current)
		case 1:
			message.Value = string(current)
		default:
			message.Complement = append(message.Complement, string(current))
		}
	}

	for {
		if _, err := os.Stdin.Read(in); err != nil {
			return nil, err
		}

		switch in[0] {
		case interactiveFieldSeparator:
			assignCurrent()
			current = nil
			currentField++
		case interactiveEndOfCommand:
			if current != nil {
				assignCurrent()
			}
			return message, nil
		default:
			current = append(current, in[0])
		}
	}
}

func (c *client) readIncomingMessages() {
	for {
		msg, err := c.client.Read()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			if errors.Is(err, io.EOF) {
				c.printString("Server closed the connection")
			} else {
				c.interactivePrintError("Failed to read response", err)
			}

			close(c.stopped)

			return
		}

		c.interactivePrintResponse(msg)
	}
}
