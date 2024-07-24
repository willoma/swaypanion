package main

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/signal"

	"github.com/willoma/swaypanion/socket"
	socketclient "github.com/willoma/swaypanion/socket/client"
)

const (
	cliFieldSeparator = ':'
	cliEndOfCommand   = '\n'
)

type cli struct {
	client *socketclient.Client

	stopped chan struct{}
}

func runInteractive(client *socketclient.Client) {
	c := &cli{client: client, stopped: make(chan struct{})}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	go c.readResponse()
	go c.handleCLI()

	select {
	case <-ctx.Done():
	case <-c.stopped:
	}
}

func (c *cli) handleCLI() {
	for {
		msg, err := c.prompt()
		if err != nil {
			if errors.Is(err, io.EOF) {
				newLine()
				close(c.stopped)
				return
			}

			printError("Failed to get instruction", err)
		}

		switch msg.Command {
		case "q", "qu", "qui", "quit", "e", "ex", "exi", "exit":
			close(c.stopped)
			return
		case "help":
			printLine("exit: Close the swaypanion client")
			printLine("quit: Close the swaypanion client")
		}

		if err := c.client.Send(msg); err != nil {
			printError("Failed to send command", err)
		}
	}
}

func (c *cli) readResponse() {
	for {
		msg, err := c.client.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				printLine("Server closed the connection")
				close(c.stopped)
			} else if !errors.Is(err, net.ErrClosed) {
				printError("Error reading response", err)
				close(c.stopped)
			}

			return
		}

		printResponse(msg)
		printPrompt()
	}
}

func (c *cli) prompt() (*socket.Message, error) {
	printPrompt()

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
		case cliFieldSeparator:
			assignCurrent()
			current = nil
			currentField++
		case cliEndOfCommand:
			if current != nil {
				assignCurrent()
			}
			return message, nil
		default:
			current = append(current, in[0])
		}
	}
}
