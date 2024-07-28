package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/signal"
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
	stdin := bufio.NewReader(os.Stdin)

	for {
		c.interactivePrintPrompt()

		msg, err := readMessage(stdin)
		if err != nil {
			if errors.Is(err, io.EOF) {
				os.Stdout.Write([]byte{'\n'})
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
			c.printLine("Ctrl-D: Close the swaypanion client")
			c.printLine("exit: Close the swaypanion client")
			c.printLine("quit: Close the swaypanion client")
		}

		if err := c.client.Send(msg); err != nil {
			c.interactivePrintError("Failed to send command", err)
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
				c.printLine("Server closed the connection")
			} else {
				c.interactivePrintError("Failed to read response", err)
			}

			close(c.stopped)

			return
		}

		c.interactivePrintResponse(msg)
	}
}
